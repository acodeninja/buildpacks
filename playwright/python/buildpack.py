import os
import sys
from pathlib import Path

try:
    from tomllib import loads
except ImportError as e:
    from pip._vendor.tomli import loads

data = loads(Path(os.environ.get("CNB_BUILDPACK_DIR")).joinpath("buildpack.toml").read_text())

print(sys.argv)
