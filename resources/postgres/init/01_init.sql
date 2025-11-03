-- SCHEMA の作成
CREATE SCHEMA IF NOT EXISTS tsudzuri;

-- User テーブル (tsudzuri.users)
CREATE TABLE
    IF NOT EXISTS tsudzuri.users (
        id SERIAL PRIMARY KEY,
        uid VARCHAR(255) UNIQUE NOT NULL,
        provider VARCHAR(20) NOT NULL,
        email VARCHAR(255) UNIQUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

COMMENT ON TABLE tsudzuri.users IS 'ユーザー情報を管理するテーブル。匿名ユーザーと登録ユーザーの両方を格納';

COMMENT ON COLUMN tsudzuri.users.id IS 'プライマリーキー';

COMMENT ON COLUMN tsudzuri.users.uid IS 'ユーザーの識別子';

COMMENT ON COLUMN tsudzuri.users.provider IS 'アカウントの永続化に使用された認証プロバイダ（例: anonymous, google.com, facebook.com）';

COMMENT ON COLUMN tsudzuri.users.email IS '登録ユーザーのメールアドレス';

COMMENT ON COLUMN tsudzuri.users.created_at IS 'レコード作成日時';

COMMENT ON COLUMN tsudzuri.users.updated_at IS 'レコード更新日時';

CREATE INDEX IF NOT EXISTS idx_users_uid ON tsudzuri.users (uid);

-- Pages (綴り) テーブル (tsudzuri.pages)
CREATE TABLE
    IF NOT EXISTS tsudzuri.pages (
        id SERIAL PRIMARY KEY,
        title VARCHAR(50) NOT NULL,
        creator_id INTEGER NOT NULL REFERENCES tsudzuri.users (id) ON DELETE CASCADE,
        invite_code VARCHAR(8) UNIQUE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

COMMENT ON TABLE tsudzuri.pages IS '「綴り」と呼ばれるリンク集を管理するテーブル。複数人で共同編集可能';

COMMENT ON COLUMN tsudzuri.pages.id IS 'プライマリーキー';

COMMENT ON COLUMN tsudzuri.pages.title IS '綴りのタイトル（最大50文字）';

COMMENT ON COLUMN tsudzuri.pages.creator_id IS '作成者のユーザーID';

COMMENT ON COLUMN tsudzuri.pages.invite_code IS '8文字の招待コード。共同編集者の招待に使用';

COMMENT ON COLUMN tsudzuri.pages.created_at IS 'レコード作成日時';

COMMENT ON COLUMN tsudzuri.pages.updated_at IS 'レコード更新日時';

CREATE INDEX IF NOT EXISTS idx_pages_creator ON tsudzuri.pages (creator_id);

-- Link items テーブル (tsudzuri.link_items)
CREATE TABLE
    IF NOT EXISTS tsudzuri.link_items (
        id SERIAL PRIMARY KEY,
        page_id INTEGER NOT NULL REFERENCES tsudzuri.pages (id) ON DELETE CASCADE,
        url TEXT NOT NULL,
        memo TEXT,
        priority INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

COMMENT ON TABLE tsudzuri.link_items IS '綴りに含まれる個々のリンクとそのメタデータを管理するテーブル';

COMMENT ON COLUMN tsudzuri.link_items.id IS 'プライマリーキー';

COMMENT ON COLUMN tsudzuri.link_items.page_id IS '所属する綴り（page）のID';

COMMENT ON COLUMN tsudzuri.link_items.url IS 'リンクのURL';

COMMENT ON COLUMN tsudzuri.link_items.memo IS 'ユーザーが付けたメモ';

COMMENT ON COLUMN tsudzuri.link_items.priority IS 'リンクの表示順序（同一page_id内での並び順）';

COMMENT ON COLUMN tsudzuri.link_items.created_at IS 'レコード作成日時';

COMMENT ON COLUMN tsudzuri.link_items.updated_at IS 'レコード更新日時';

CREATE INDEX IF NOT EXISTS idx_link_items_page_order ON tsudzuri.link_items (page_id, order_index);