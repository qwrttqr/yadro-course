BEGIN;

INSERT INTO words (comics_id, word)
SELECT c.id, w.word
FROM comics AS c
         CROSS JOIN LATERAL unnest(c.words) WITH ORDINALITY AS w(word, position)
ORDER BY c.id, w.position;

ALTER TABLE comics DROP COLUMN words;

COMMIT;