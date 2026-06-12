#!/usr/bin/env python3
"""One-time migration: convert markdown chapter content to HTML."""
import os
import re
import markdown
import psycopg2

DATABASE_URL = os.environ.get("DATABASE_URL", 
    "postgresql://postgres:***@fnnas.local:5433/novelist")

def is_html(text):
    """Check if content is already HTML."""
    stripped = text.strip()
    return stripped.startswith("<") and (
        "<p>" in stripped or "<h" in stripped or "<pre>" in stripped or 
        "<ul>" in stripped or "<ol>" in stripped or "<blockquote>" in stripped or
        "<hr" in stripped or "<div>" in stripped
    )

def convert_md_to_html(md_text):
    """Convert markdown to HTML, preserving structure."""
    html = markdown.markdown(md_text, extensions=["fenced_code", "codehilite", "tables"])
    return html

def main():
    conn = psycopg2.connect(DATABASE_URL)
    cur = conn.cursor()
    
    # Find chapters with content
    cur.execute("SELECT id, content FROM chapters WHERE content IS NOT NULL AND content != ''")
    rows = cur.fetchall()
    
    converted = 0
    skipped = 0
    errors = 0
    
    for chapter_id, content in rows:
        if is_html(content):
            skipped += 1
            continue
        
        try:
            html = convert_md_to_html(content)
            cur.execute("UPDATE chapters SET content = %s WHERE id = %s", (html, chapter_id))
            converted += 1
        except Exception as e:
            print(f"Error converting chapter {chapter_id}: {e}")
            errors += 1
    
    conn.commit()
    cur.close()
    conn.close()
    
    print(f"Migration complete: {converted} converted, {skipped} skipped (already HTML), {errors} errors")

if __name__ == "__main__":
    main()
