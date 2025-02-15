CREATE TABLE months (
        id VARCHAR(50) PRIMARY KEY ,
        name VARCHAR(255),
        code VARCHAR(255),
        storage_id VARCHAR(50),
        type VARCHAR(50),
        parent_id VARCHAR(50),
        deleted_type VARCHAR(50),
        global_content_version VARCHAR(50),
        file_id VARCHAR(50),
        size VARCHAR(50),
        create_time VARCHAR(255),
        update_time VARCHAR(255),
        delete_time VARCHAR(255),
        created_by VARCHAR(50),
        updated_by VARCHAR(50),
        deleted_by VARCHAR(50),
        download_url TEXT,
        detail_url TEXT
);

CREATE TABLE CallInfo (
        id VARCHAR(100) PRIMARY KEY ,
        portal_user_id VARCHAR(50),
        portal_number VARCHAR(50),
        phone_number VARCHAR(50),
        call_id VARCHAR(255),
        external_call_id VARCHAR(255) NULL,
        call_category VARCHAR(50),
        call_duration VARCHAR(100),
        call_start_date TIMESTAMP,
        call_record_url TEXT,
        call_vote VARCHAR(100),
        cost VARCHAR(100),
        cost_currency VARCHAR(10),
        call_failed_code VARCHAR(50),
        call_failed_reason TEXT NULL,
        crm_entity_type VARCHAR(50),
        crm_entity_id VARCHAR(50),
        crm_activity_id VARCHAR(50),
        rest_app_id VARCHAR(50),
        rest_app_name VARCHAR(255),
        transcript_id VARCHAR(50) NULL,
        transcript_pending VARCHAR(10),
        session_id VARCHAR(50) NULL,
        redial_attempt VARCHAR(100),
        comment TEXT NULL,
        record_duration VARCHAR(100),
        record_file_id VARCHAR(100),
        call_type VARCHAR(100)
);

CREATE TABLE users (
        id VARCHAR(50) PRIMARY KEY ,
        xml_id VARCHAR(50),
        active VARCHAR(50),
        name VARCHAR(100),
        last_name VARCHAR(100),
        second_name VARCHAR(100),
        email VARCHAR(255),
        last_login VARCHAR(255),
        time_zone VARCHAR(50),
        time_zone_offset INT,
        personal_photo TEXT,
        personal_gender VARCHAR(10),
        personal_www TEXT,
        personal_birthday VARCHAR(255),
        personal_mobile VARCHAR(20),
        personal_city VARCHAR(100),
        work_phone VARCHAR(20),
        work_position VARCHAR(255),
        uf_employment_date VARCHAR(255),
        user_type VARCHAR(50),
        department_ids TEXT
);


CREATE TABLE total (
        audio_path TEXT NOT NULL,
        call_id VARCHAR(50) REFERENCES CallInfo(id),
        user_id VARCHAR(50) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS audio_log (
                                         last_downloaded_file_id VARCHAR(255) NOT NULL PRIMARY KEY
    );

