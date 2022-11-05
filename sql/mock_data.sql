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
