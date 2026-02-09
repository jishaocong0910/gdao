CREATE TABLE test_table
(
    `auto_increment`     bigint PRIMARY KEY AUTO_INCREMENT COMMENT 'bigint auto_increment',
    `bit`                bit,
    `tinyint`            tinyint NOT NULL DEFAULT 1 COMMENT 'tinyint',
    `tinyint_unsigned`   tinyint UNSIGNED COMMENT 'tinyint unsigned',
    `tinyint_1`          tinyint(1) COMMENT 'tinyint(1)',
    `smallint`           smallint COMMENT 'smallint',
    `smallint_unsigned`  smallint UNSIGNED COMMENT 'smallint unsigned',
    `mediumint`          mediumint COMMENT 'mediumint',
    `mediumint_unsigned` mediumint UNSIGNED COMMENT 'mediumint unsigned',
    `int`                int COMMENT 'int',
    `int_unsigned`       int UNSIGNED COMMENT 'int unsigned',
    `bigint`             bigint COMMENT 'bigint',
    `bigint_unsigned`    bigint UNSIGNED COMMENT 'bigint unsigned',
    `double` double COMMENT 'double',
    `decimal`            decimal COMMENT 'decimal',
    `varchar`            varchar(255) COMMENT 'varchar(255)',
    `char`               char COMMENT 'char',
    `text`               text COMMENT 'text',
    `tinytext`           tinytext COMMENT 'tinytext',
    `mediumtext`         mediumtext COMMENT 'mediumtext',
    `longtext`           longtext COMMENT 'longtext',
    `enum`               enum ('a', 'b', 'c') COMMENT 'enum(''a'',''b'',''c'')',
    `json`               json COMMENT 'json',
    `set` set ('a', 'b', 'c') COMMENT 'set(''a'',''b'',''c'')',
    `date`               date COMMENT 'date',
    `datetime`           datetime(3) COMMENT 'datetime(3) ',
    `timestamp`          timestamp COMMENT 'timestamp',
    `year` year COMMENT 'year',
    `binary`             binary COMMENT 'binary',
    `varbinary`          varbinary(255) COMMENT 'varbinary(255)',
    `geometry`           geometry COMMENT 'geometry',
    `blob`               blob COMMENT 'blob',
    `tinyblob`           tinyblob COMMENT 'tinyblob',
    `mediumblob`         mediumblob COMMENT 'mediumblob',
    `longblob`           longblob COMMENT 'longblob',
    `numeric`            numeric COMMENT 'numeric',
    `integer`            integer COMMENT 'integer',
    `float`              float COMMENT 'float',
    `real`               real COMMENT 'real',
    `boolean`            boolean COMMENT 'boolean',
    `point`              point COMMENT 'point',
    `time`               time,
    `other`              int,
    `other2`             int
) COMMENT 'mysql';

CREATE TABLE test_table_logical_del_mode1
(
    `id`    bigint PRIMARY KEY AUTO_INCREMENT,
    `name`  varchar(255),
    `valid` char(1) NOT NULL DEFAULT '1'
);

CREATE TABLE test_table_logical_del_mode2
(
    `id`      bigint PRIMARY KEY AUTO_INCREMENT,
    `name`    varchar(255),
    `deleted` tinyint(1) NOT NULL DEFAULT 0
);