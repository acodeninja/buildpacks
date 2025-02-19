from pathlib import Path

def test_creating_a_file():
    Path('.pytest-ran').write_text("test")
    assert Path('.pytest-ran').exists()
