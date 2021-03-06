CREATE TABLE prj_user (
    userid CHARACTER VARYING(32) NOT NULL,
    nameuser CHARACTER VARYING(300) NOT NULL UNIQUE,
    chatid CHARACTER VARYING(32) NOT NULL UNIQUE,

    CONSTRAINT pk_prj_user PRIMARY KEY (userid)
)
WITH(
	OIDS=FALSE
);

ALTER TABLE prj_user OWNER TO postgres;
COMMENT ON TABLE prj_user IS 'Таблица пользователей';
COMMENT ON TABLE prj_user.userid IS 'Идентификатор пользователя';
COMMENT ON TABLE ref_user.nameuser IS 'Имя пользовтаеля в tlg';


CREATE TABLE prj_code(
    codeid CHARACTER VARYING(32) NOT NULL,
    code CHARACTER VARYING(300) NOT NULL UNIQUE,
    
    CONSTRAINT pk_prj_code PRIMARY KEY (codeid)
)
WITH(
	OIDS=FALSE
);
ALTER TABLE prj_code OWNER TO postgres;
COMMENT ON TABLE prj_code IS 'Таблица кодовых слов';
COMMENT ON TABLE prj_code.codeid IS 'Идентификатор кодового слова';
COMMENT ON TABLE prj_code.code IS 'Кодовое слово';

CREATE TABLE ref_usercode(
    keyid CHARACTER VARYING(32) NOT NULL UNIQUE, -- идентификатор 
    codeid CHARACTER VARYING(32) , -- идентификатор кодового слова
    userid CHARACTER VARYING(32) NOT NULL UNIQUE, -- идентификатор пользователя

    CONSTRAINT pk_ref_usercode PRIMARY KEY (keyid),
    CONSTRAINT pk_prj_user FOREIGN KEY (userid)
        REFERENCES prj_user (userid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT pk_prj_code FOREIGN KEY (codeid)
        REFERENCES prj_code (codeid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE RESTRICT
)
WITH(
	OIDS=FALSE
);

ALTER TABLE ref_usercode OWNER TO postgres;
COMMENT ON TABLE ref_usercode IS 'Связь между кодовым словом и пользователем';
COMMENT ON TABLE ref_usercode.keyid IS 'Идентификатор связи кодового слова и пользователя';
COMMENT ON TABLE ref_usercode.codeid IS 'Идентификатор кодового слова';
COMMENT ON TABLE ref_usercode.userid IS 'Идентификатор пользователя';




CREATE TABLE prj_letter(
    letterid CHARACTER VARYING(32) NOT NULL,
    codeid CHARACTER VARYING(32) NOT NULL,
    userid CHARACTER VARYING(32) NOT NULL,
    letter text NOT NULL,
    dataletter TIMESTAMP WITHOUT TIME ZONE,

    CONSTRAINT pk_prj_letter PRIMARY KEY (letterid),
    CONSTRAINT pk_prj_user FOREIGN KEY (userid)
        REFERENCES prj_user (userid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT pk_prj_code FOREIGN KEY (codeid)
        REFERENCES prj_code (codeid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE RESTRICT
)
WITH(
	OIDS=FALSE
);

ALTER TABLE ref_usercode OWNER TO postgres;
COMMENT ON TABLE prj_letter IS 'Таблица сообщений';
COMMENT ON TABLE prj_letter.letterid IS 'Идентификатор сообщения';
COMMENT ON TABLE prj_letter.codeid IS 'Идентификатор кодового слова';
COMMENT ON TABLE prj_letter.userid IS 'Идентификатор пользователя';
COMMENT ON TABLE prj_letter.letter IS 'Сообщение';
COMMENT ON TABLE prj_letter.dataletter IS 'Дата создания сообщения';



CREATE TABLE prj_botwork(
    botworkid CHARACTER VARYING(32) NOT NULL,
    userid CHARACTER VARYING(32) NOT NULL UNIQUE,
    botworkflag boolean NOT NULL,

    CONSTRAINT pk_prj_botworkid PRIMARY KEY (botworkid),
    CONSTRAINT pk_prj_user FOREIGN KEY (userid)
        REFERENCES prj_user (userid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE CASCADE
)
WITH(
	OIDS=FALSE
);

CREATE TABLE prj_lastusercommand(
    commandid CHARACTER VARYING(32) NOT NULL,
    userid CHARACTER VARYING(32) NOT NULL,
    command CHARACTER VARYING(32) NOT NULL,
    datacommand TIMESTAMP WITHOUT TIME ZONE,

    CONSTRAINT pk_prj_commandid PRIMARY KEY (commandid),
        CONSTRAINT pk_prj_user FOREIGN KEY (userid)
        REFERENCES prj_user (userid) MATCH SIMPLE 
        ON UPDATE RESTRICT ON DELETE CASCADE
)
WITH(
	OIDS=FALSE
);