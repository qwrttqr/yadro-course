BEGIN;

ALTER TABLE comics
    ADD COLUMN words text[] NOT NULL DEFAULT '{}';

UPDATE comics AS c
SET words = restored.words
FROM (SELECT comics_id, array_agg(word ORDER BY id) AS words
      FROM words
      GROUP BY comics_id) AS restored
WHERE restored.comics_id = c.id;

DELETE
FROM words;

COMMIT;