CREATE TABLE IF NOT EXISTS energyCounter (
    Date TEXT NOT NULL,
    EnergyKwh REAL NOT NULL,
    PRIMARY KEY (Date DESC)
);

CREATE TABLE IF NOT EXISTS waterCounter (
    Date TEXT NOT NULL,
    ColdWaterLiters INT NOT NULL,
    HotWaterLiters INT NOT NULL,
    PRIMARY KEY (Date DESC)
);

CREATE TABLE IF NOT EXISTS bankTransactions (
    TransactionId INTEGER PRIMARY KEY AUTOINCREMENT,
    AccountNumber TEXT NOT NULL,
    ExecutionDate TEXT NOT NULL,
    OrderDate TEXT NOT NULL,
    TType TEXT NULL,
    AmountCurrency TEXT NOT NULL,
    AmountValue REAL NOT NULL,
    EndingBalanceCurrency TEXT NULL,
    EndingBalanceValue REAL NULL,
    Description TEXT NOT NULL,

    UNIQUE(AccountNumber, ExecutionDate, OrderDate, AmountValue, Description)
);

CREATE TABLE IF NOT EXISTS documents (
    DocumentId INT NOT NULL,
    DocumentName TEXT NOT NULL,
    UploadDate TEXT NOT NULL,
    DocumentDate TEXT NULL,
    Category TEXT NOT NULL,
    PersonInvolved TEXT NULL,
    FileExtension TEXT NOT NULL,
    FileSize INT NOT NULL,

    PRIMARY KEY (DocumentId),
    UNIQUE(DocumentName, DocumentDate, Category, PersonInvolved)
);

CREATE TABLE IF NOT EXISTS documentFiles (
    DocumentId INT NOT NULL,
    FileBytes BLOB NOT NULL
);
