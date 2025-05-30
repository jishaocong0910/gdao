CREATE TABLE oracle
(
    "char"                           CHAR(10),
    "char_2"                         CHAR(1 BYTE),
    "char_3"                         CHAR(1 CHAR),
    "varchar2"                       VARCHAR2(10),
    "varchar2_2"                     VARCHAR2(1 BYTE),
    "varchar2_3"                     VARCHAR2(1 CHAR),
    "varchar1"                       VARCHAR(10),
    "varchar1_2"                     VARCHAR(1 BYTE),
    "varchar1_3"                     VARCHAR(2 CHAR),
    "clob"                           CLOB,
    "nclob"                          NCLOB,
    "nchar"                          NCHAR(10),
    "nvarchar2"                      NVARCHAR2(10),
    "number"                         NUMBER(10),
    "number_2"                       NUMBER(10, 0),
    "number_3"                       NUMBER(10, 4),
    "float"                          FLOAT,
    "binary_float"                   BINARY_FLOAT,
    "binary_double"                  BINARY_DOUBLE,
    "date"                           DATE,
    "timestamp"                      TIMESTAMP,
    "timestamp_with_time_zone"       TIMESTAMP WITH TIME ZONE,
    "timestamp_with_local_time_zone" TIMESTAMP WITH LOCAL TIME ZONE,
    "blob"                           BLOB,
    "raw"                            RAW(10),
    "long_raw"                       LONG RAW,
    "rowid"                          ROWID,
    "urowid"                         UROWID
);

COMMENT ON TABLE oracle IS 'oracle';
COMMENT ON COLUMN oracle."char_2" IS 'char_2 ';
COMMENT ON COLUMN oracle."char_3" IS 'char_3 ';
COMMENT ON COLUMN oracle."varchar2" IS 'varchar2 ';
COMMENT ON COLUMN oracle."varchar2_2" IS 'varchar2_2 ';
COMMENT ON COLUMN oracle."varchar2_3" IS 'varchar2_3 ';
COMMENT ON COLUMN oracle."varchar1" IS 'varchar1 ';
COMMENT ON COLUMN oracle."varchar1_2" IS 'varchar1_2 ';
COMMENT ON COLUMN oracle."varchar1_3" IS 'varchar1_3 ';
COMMENT ON COLUMN oracle."clob" IS 'clob ';
COMMENT ON COLUMN oracle."nclob" IS 'nclob ';
COMMENT ON COLUMN oracle."nchar" IS 'nchar ';
COMMENT ON COLUMN oracle."nvarchar2" IS 'nvarchar2 ';
COMMENT ON COLUMN oracle."number" IS 'number ';
COMMENT ON COLUMN oracle."number_2" IS 'number_2 ';
COMMENT ON COLUMN oracle."number_3" IS 'number_3 ';
COMMENT ON COLUMN oracle."float" IS 'float ';
COMMENT ON COLUMN oracle."binary_float" IS 'binary_float ';
COMMENT ON COLUMN oracle."binary_double" IS 'binary_double ';
COMMENT ON COLUMN oracle."date" IS 'date ';
COMMENT ON COLUMN oracle."timestamp" IS 'timestamp ';
COMMENT ON COLUMN oracle."timestamp_with_time_zone" IS 'timestamp_with_time_zone ';
COMMENT ON COLUMN oracle."timestamp_with_local_time_zone" IS 'timestamp_with_local_time_zone ';
COMMENT ON COLUMN oracle."blob" IS 'blob ';
COMMENT ON COLUMN oracle."raw" IS 'raw ';
COMMENT ON COLUMN oracle."long_raw" IS 'long_raw ';
COMMENT ON COLUMN oracle."rowid" IS 'rowid ';