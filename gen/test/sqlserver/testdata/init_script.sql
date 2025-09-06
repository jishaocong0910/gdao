CREATE TABLE test_table
(
    tinyint          tinyint,
    smallint         smallint,
    int              int IDENTITY(4,2),
    bigint           bigint,
    bit              bit,
    decimal          decimal,
    numeric          numeric,
    money            money,
    smallmoney       smallmoney,
    float            float,
    real             real,
    date             date,
    time             time,
    datetime2        datetime2,
    datetimeoffset   datetimeoffset,
    datetime         datetime,
    smalldatetime    smalldatetime,
    char             char,
    varchar          varchar,
    text             text,
    nchar            nchar,
    nvarchar         nvarchar,
    ntext            ntext,
    binary           binary,
    varbinary        varbinary,
    image            image,
    geography        geography,
    geometry         geometry,
    hierarchyid      hierarchyid,
    uniqueidentifier uniqueidentifier,
    xml              xml
);

EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'sqlserver', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'smallint' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'smallint';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'int' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'int';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'bigint' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'bigint';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'bit' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'bit';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'decimal' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'decimal';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'numeric' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'numeric';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'money' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'money';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'smallmoney' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'smallmoney';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'float' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'float';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'real' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'real';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'date' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'date';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'time' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'time';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'datetime2' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'datetime2';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'datetimeoffset' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'datetimeoffset';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'datetime' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'datetime';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'smalldatetime' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'smalldatetime';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'char' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'char';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'varchar' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'varchar';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'text' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'text';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'nchar' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'nchar';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'nvarchar' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'nvarchar';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'ntext' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'ntext';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'binary' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'binary';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'varbinary' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'varbinary';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'image' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'image';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'geography' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'geography';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'geometry' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'geometry';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'hierarchyid' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'hierarchyid';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'uniqueidentifier' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'uniqueidentifier';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'xml' , @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'test_table', @level2type=N'COLUMN', @level2name=N'xml';


