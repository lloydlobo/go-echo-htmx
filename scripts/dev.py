#!/usr/bin/env python3

"""
scripts/dev.py

Pre-requisites:

    - Ensure "tailwindcss" standalone CLI and "air" executables are installed

Run:

    $ python3 scripts/dev.py

Make the script executable:

    $ chmod +x scripts/dev.py
    $ ./scripts/dev.py
"""

from concurrent.futures import ThreadPoolExecutor
import logging, os, subprocess

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def run_tailwindcss(in_filepath, out_filepath, with_watch=True):
    cmd = [
        "tailwindcss",
        "-i",
        in_filepath,
        "-o",
        out_filepath,
        "--minify",
        "--watch" if with_watch else "",
    ]
    try:
        result = subprocess.run(cmd, text=True, check=True)
        logger.info("$ %s\n%s", " ".join(cmd), result.stdout)
    except subprocess.CalledProcessError as e:
        logger.error(f"Error running Tailwind {''.join(cmd)}: {e}")
        raise


def run_air():
    try:
        cmd = ["air"]
        result = subprocess.run(cmd, shell=True, text=True, check=True)
        logger.info("$ %s\n%s", " ".join(cmd), result.stdout)
    except subprocess.CalledProcessError as e:
        logger.error("Error running Air: %s", e)
        raise
    finally:
        if str(result.returncode) == "0":
            return
        logger.error(f"error while running {''.join(cmd)}: {result.returncode}")
        raise


def main():
    is_parallel = True

    root_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))
    static_dir = os.path.join(root_dir, "static")
    templates_dir = os.path.join(root_dir, "templates")
    templates_css_dir = os.path.join(templates_dir, "css")
    in_css_filepath = os.path.join(templates_css_dir, "globals.css")
    out_css_filepath = os.path.join(static_dir, "css", "style.css")

    if is_parallel:
        with ThreadPoolExecutor(max_workers=2) as executor:
            executor.submit(run_tailwindcss, in_css_filepath, out_css_filepath)
            executor.submit(run_air)
    else:
        run_tailwindcss(in_css_filepath, out_css_filepath, with_watch=False)
        run_air()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nReceived KeyboardInterrupt. Exiting gracefully.")
