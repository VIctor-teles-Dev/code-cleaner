-- Conteúdo do blog por idioma. Slug/published_at continuam globais (um post,
-- um slug, várias traduções). Title/content saem de posts e viram traduções.
CREATE TABLE post_translations (
    post_id BIGINT NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    locale  TEXT NOT NULL,
    title   TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (post_id, locale)
);

INSERT INTO post_translations (post_id, locale, title, content)
    SELECT id, 'pt-BR', title, content FROM posts;

ALTER TABLE posts DROP COLUMN title, DROP COLUMN content;

-- Nome das tags por idioma (o slug continua único e global).
CREATE TABLE tag_translations (
    tag_id BIGINT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    name   TEXT NOT NULL,
    PRIMARY KEY (tag_id, locale)
);

INSERT INTO tag_translations (tag_id, locale, name)
    SELECT id, 'pt-BR', name FROM tags;

ALTER TABLE tags DROP COLUMN name;
