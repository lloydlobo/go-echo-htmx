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


def get_root_dir():
    return os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))


def get_static_dir(root_dir):
    return os.path.join(root_dir, "static")


def get_templates_dir(root_dir):
    return os.path.join(root_dir, "templates")


def get_templates_css_dir(dir):
    return os.path.join(dir, "css")


def get_input_css_filepath(dir):
    return os.path.join(dir, "globals.css")


def get_output_css_filepath(static_dir):
    return os.path.join(static_dir, "css", "style.css")


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
        logger.error("Error running Tailwind CSS: %s", e)
        raise


def run_air():
    subprocess.run("air", shell=True, text=True, check=True)


def main():
    is_parallel = True

    root_dir = get_root_dir()
    static_dir = get_static_dir(root_dir)
    templates_dir = get_templates_dir(root_dir)
    templates_css_dir = get_templates_css_dir(dir=templates_dir)
    input_css_filepath = get_input_css_filepath(dir=templates_css_dir)
    output_css_filepath = get_output_css_filepath(static_dir)
    print(f"{input_css_filepath, output_css_filepath}")

    if is_parallel:
        with ThreadPoolExecutor(max_workers=2) as executor:
            executor.submit(run_tailwindcss, input_css_filepath, output_css_filepath)
            executor.submit(run_air)
    else:
        run_tailwindcss(input_css_filepath, output_css_filepath, with_watch=False)
        run_air()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nReceived KeyboardInterrupt. Exiting gracefully.")
