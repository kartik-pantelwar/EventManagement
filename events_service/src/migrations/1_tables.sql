-- Events table to store event information
CREATE TABLE IF NOT EXISTS events (
    event_id SERIAL PRIMARY KEY,
    event_name TEXT NOT NULL,
    organizer_id INTEGER NOT NULL,
    place TEXT NOT NULL,
    event_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    filled INTEGER DEFAULT 0 CHECK (filled >= 0 AND filled <= capacity),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

-- Constraint: Only one event per place at a time (no overlapping times)
CONSTRAINT check_end_time_after_start CHECK (end_time > start_time)
);

-- User bookings table (renamed to match requirement: userbooked_events)
CREATE TABLE IF NOT EXISTS userbooked_events (
    booking_id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events (event_id) ON DELETE CASCADE,
    cid INTEGER NOT NULL, -- customer id
    cemail TEXT NOT NULL,
    cusername TEXT NOT NULL,
    booked_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (cid, event_id) -- One customer can only book one event once
);

-- Index for better query performance
CREATE INDEX IF NOT EXISTS idx_events_date_time ON events (
    event_date,
    start_time,
    end_time
);

CREATE INDEX IF NOT EXISTS idx_events_organizer ON events (organizer_id);

CREATE INDEX IF NOT EXISTS idx_events_place ON events (place);

CREATE INDEX IF NOT EXISTS idx_userbooked_events_customer ON userbooked_events (cid);

CREATE INDEX IF NOT EXISTS idx_userbooked_events_event ON userbooked_events (event_id);

-- Function to check if a place is available for the given time slot
CREATE OR REPLACE FUNCTION check_place_availability(
    p_place TEXT,
    p_event_date DATE,
    p_start_time TIME,
    p_end_time TIME,
    p_exclude_event_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    conflict_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO conflict_count
    FROM events
    WHERE place = p_place
      AND event_date = p_event_date
      AND (
          (start_time <= p_start_time AND end_time > p_start_time) OR
          (start_time < p_end_time AND end_time >= p_end_time) OR
          (start_time >= p_start_time AND end_time <= p_end_time)
      )
      AND (p_exclude_event_id IS NULL OR event_id != p_exclude_event_id);
    
    RETURN conflict_count = 0; -- Return true if no conflicts (place is available)
END;
$$ LANGUAGE plpgsql;

-- Function to check customer time conflicts
CREATE OR REPLACE FUNCTION check_customer_time_conflict(
    p_customer_id INTEGER,
    p_event_date DATE,
    p_start_time TIME,
    p_end_time TIME,
    p_exclude_event_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    conflict_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO conflict_count
    FROM userbooked_events ub
    JOIN events e ON ub.event_id = e.event_id
    WHERE ub.cid = p_customer_id
      AND e.event_date = p_event_date
      AND (
          (e.start_time <= p_start_time AND e.end_time > p_start_time) OR
          (e.start_time < p_end_time AND e.end_time >= p_end_time) OR
          (e.start_time >= p_start_time AND e.end_time <= p_end_time)
      )
      AND (p_exclude_event_id IS NULL OR e.event_id != p_exclude_event_id);
    
    RETURN conflict_count > 0; -- Return true if there's a conflict
END;
$$ LANGUAGE plpgsql;

-- Trigger to update filled count when bookings are added/removed
CREATE OR REPLACE FUNCTION update_event_filled_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE events 
        SET filled = filled + 1 
        WHERE event_id = NEW.event_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE events 
        SET filled = filled - 1 
        WHERE event_id = OLD.event_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
DROP TRIGGER IF EXISTS trigger_update_event_filled_insert ON userbooked_events;

CREATE TRIGGER trigger_update_event_filled_insert
    AFTER INSERT ON userbooked_events
    FOR EACH ROW EXECUTE FUNCTION update_event_filled_count();

DROP TRIGGER IF EXISTS trigger_update_event_filled_delete ON userbooked_events;

CREATE TRIGGER trigger_update_event_filled_delete
    AFTER DELETE ON userbooked_events
    FOR EACH ROW EXECUTE FUNCTION update_event_filled_count();

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_events_updated_at ON events;

CREATE TRIGGER trigger_update_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- User bookings table to track which users have joined which events
CREATE TABLE IF NOT EXISTS user_bookings (
    booking_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL REFERENCES events (event_id) ON DELETE CASCADE,
    booked_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, event_id)
);

-- Index for better query performance
CREATE INDEX IF NOT EXISTS idx_events_date_time ON events (event_date, event_time);

CREATE INDEX IF NOT EXISTS idx_events_organizer ON events (organizer_id);

CREATE INDEX IF NOT EXISTS idx_events_venue ON events (event_venue);

CREATE INDEX IF NOT EXISTS idx_user_bookings_user ON user_bookings (user_id);

CREATE INDEX IF NOT EXISTS idx_user_bookings_event ON user_bookings (event_id);

-- Function to check for time conflicts for users
CREATE OR REPLACE FUNCTION check_user_time_conflict(
    p_user_id INTEGER,
    p_event_date DATE,
    p_event_time TIME,
    p_exclude_event_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    conflict_count INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO conflict_count
    FROM user_bookings ub
    JOIN events e ON ub.event_id = e.event_id
    WHERE ub.user_id = p_user_id
      AND e.event_date = p_event_date
      AND e.event_time = p_event_time
      AND (p_exclude_event_id IS NULL OR e.event_id != p_exclude_event_id);
    
    RETURN conflict_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update current_count when bookings are added/removed
CREATE OR REPLACE FUNCTION update_event_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE events 
        SET current_count = current_count + 1 
        WHERE event_id = NEW.event_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE events 
        SET current_count = current_count - 1 
        WHERE event_id = OLD.event_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
DROP TRIGGER IF EXISTS trigger_update_event_count_insert ON user_bookings;

CREATE TRIGGER trigger_update_event_count_insert
    AFTER INSERT ON user_bookings
    FOR EACH ROW EXECUTE FUNCTION update_event_count();

DROP TRIGGER IF EXISTS trigger_update_event_count_delete ON user_bookings;

CREATE TRIGGER trigger_update_event_count_delete
    AFTER DELETE ON user_bookings
    FOR EACH ROW EXECUTE FUNCTION update_event_count();

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_events_updated_at ON events;

CREATE TRIGGER trigger_update_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();