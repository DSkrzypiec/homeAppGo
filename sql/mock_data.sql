DELETE FROM users;
INSERT INTO users (UserId, Email, Username, PasswordHashed, Salt)
VALUES
    (
        1,
        'test@test.dev',
        'testuser',
        'e5ad54cf823a8de54b9ed452523567720cd03f3b60af5af38d81997151bdae8f',
        '21sadG4#aVB'
    )
;

DELETE FROM waterCounter;
INSERT INTO waterCounter (Date, ColdWaterLiters, HotWaterLiters)
VALUES
    ('2022-10-01', 300234, 179893),
    ('2022-10-14', 310894, 182893),
    ('2022-11-01', 315119, 186220)
;

DELETE FROM energyCounter;
INSERT INTO energyCounter (Date, EnergyKwh)
VALUES
    ('2022-10-01', 7542.4),
    ('2022-10-14', 7782.8),
    ('2022-11-01', 7991.1)
;
