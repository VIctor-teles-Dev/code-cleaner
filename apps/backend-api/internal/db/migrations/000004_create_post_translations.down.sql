-- Restaura o nome único das tags a partir da tradução pt-BR.
ALTER TABLE tags ADD COLUMN name TEXT NOT NULL DEFAULT '';
UPDATE tags t SET name = tt.name
    FROM tag_translations tt WHERE tt.tag_id = t.id AND tt.locale = 'pt-BR';
DROP TABLE tag_translations;

-- Restaura title/content dos posts a partir da tradução pt-BR.
ALTER TABLE posts ADD COLUMN title TEXT NOT NULL DEFAULT '';
ALTER TABLE posts ADD COLUMN content TEXT NOT NULL DEFAULT '';
UPDATE posts p SET title = ptr.title, content = ptr.content
    FROM post_translations ptr WHERE ptr.post_id = p.id AND ptr.locale = 'pt-BR';
DROP TABLE post_translations;
