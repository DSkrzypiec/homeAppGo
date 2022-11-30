CREATE TABLE IF NOT EXISTS users (
    UserId INT NOT NULL,
    Email TEXT NOT NULL,
    Username TEXT NOT NULL,
    PasswordHashed TEXT NOT NULL,
    Salt TEXT NOT NULL,
    IsActive INT NOT NULL DEFAULT 1,
    CreateDate TEXT NOT NULL DEFAULT CURRENT_DATE,

    PRIMARY KEY (UserId),
    UNIQUE(Username)
);

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

-- For full text search on documents
CREATE VIRTUAL TABLE IF NOT EXISTS documentsFts5 USING fts5(
    DocumentId,
    DocumentName,
    UploadDate,
    DocumentDate,
    Category,
    PersonInvolved,
    FileExtension
);

DELETE FROM documentsFts5;
INSERT INTO documentsFts5 (
    DocumentId, DocumentName, UploadDate, DocumentDate, Category,
    PersonInvolved, FileExtension
)
SELECT
    DocumentId,
    DocumentName,
    UploadDate,
    DocumentDate,
    Category,
    PersonInvolved,
    FileExtension
FROM
    documents
ORDER BY
    DocumentId
;

CREATE TABLE IF NOT EXISTS books (
    BookId INT NOT NULL,
    Title TEXT NOT NULL,
    Authors TEXT NOT NULL,
    Publisher TEXT NULL,
    PublishingYear INT NULL,
    Category TEXT NOT NULL,
    Language TEXT NOT NULL,
    FileExtension TEXT NULL,
    FileSize INT NULL,
    UploadDate TEXT NULL,

    PRIMARY KEY (BookId),
    UNIQUE(Title, Authors, Publisher, PublishingYear, Language, FileExtension)
);

CREATE TABLE IF NOT EXISTS bookFiles (
    BookId INT NOT NULL,
    FileBytes BLOB NOT NULL
);

-- For full text search on books
CREATE VIRTUAL TABLE IF NOT EXISTS booksFts5 USING fts5(
    BookId,
    Title,
    Authors,
    Publisher,
    PublishingYear,
    Category,
    Language,
    FileExtension,
    UploadDate
);
