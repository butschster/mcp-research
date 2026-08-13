-- The standalone `artifact` entry type became the `html` block of a block
-- document, so an entry can put prose, alerts and a custom visual in one place
-- instead of being a single HTML page. `artifact` stays valid on input as sugar
-- for a one-block document, but it is no longer stored.
--
-- Rewrite the rows in place: entry_type becomes 'blocks' and content becomes a
-- document holding the same HTML in one wide `html` block. The title and caption
-- are NOT extracted here — the entry already carries the title and description
-- that were derived from <title>/<meta> when it was created, and re-deriving them
-- in SQL would need an HTML parser. A missing block title only means the
-- renderer shows no caption above the frame.
--
-- json_quote() escapes the HTML for embedding in a JSON string; SQLite ships
-- JSON1 in the modernc driver, so this is available without an extension.
UPDATE entries
SET entry_type = 'blocks',
    content = '{"version":1,"blocks":[{"type":"html","data":{"html":'
              || json_quote(content)
              || ',"width":"wide"}}]}'
WHERE entry_type = 'artifact';
