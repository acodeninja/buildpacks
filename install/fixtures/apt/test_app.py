from shutil import which


def test_curl_is_installed():
    assert which("curl") == '/layers/acodeninja_install/apt-install/usr/bin/curl'


def test_git_is_installed():
    assert which("git") == '/layers/acodeninja_install/apt-install/usr/bin/git'


def test_psql_is_installed():
    assert which("psql") == '/layers/acodeninja_install/apt-install/usr/bin/psql'
