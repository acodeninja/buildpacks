import os
import re
from playwright.sync_api import Page, expect


def test_has_title(page: Page):
    page.goto("https://example.com/")

    # Expect a title "to contain" a substring.
    expect(page).to_have_title(re.compile("Example Domain"))


def test_get_started_link(page: Page):
    page.goto("https://example.com/")

    # Click the get started link.
    page.get_by_role("link", name="More information...").click()

    # Expects page to have a heading with the name of Installation.
    expect(page.get_by_role("heading", name="Example Domains")).to_be_visible()
