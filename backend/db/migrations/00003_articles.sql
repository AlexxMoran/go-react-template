-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title        TEXT        NOT NULL,
    content      TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'draft'
                 CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_articles_author_id ON articles (author_id);
CREATE INDEX idx_articles_status ON articles (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE articles;
-- +goose StatementEnd
