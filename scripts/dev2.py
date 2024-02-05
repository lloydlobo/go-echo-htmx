import logging
import os
import subprocess
import threading

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logr = logging.getLogger(__name__)


def run_cmd(name, *args):
    cmd = [name, *args]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        logr.info("$ %s\n%s", " ".join(cmd), result.stdout)
    except subprocess.CalledProcessError as e:
        logr.error("Error running command: %s", e)
        raise


def run_tw(in_path, out_path, with_watch):
    flags = ["-i", in_path, "-o", out_path, "--minify"]
    if with_watch:
        flags.append("--watch")
    run_cmd("tailwindcss", *flags)


def run_air():
    run_cmd("air")


def main():
    parallel = True

    root_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))
    static_dir = os.path.join(root_dir, "static")
    templates_dir = os.path.join(root_dir, "templates")
    templates_css_dir = os.path.join(templates_dir, "css")
    in_css_path = os.path.join(templates_css_dir, "globals.css")
    out_css_path = os.path.join(static_dir, "css", "style.css")

    if parallel:
        thread_tw = threading.Thread(
            target=run_tw, args=(in_css_path, out_css_path, True)
        )
        thread_air = threading.Thread(target=run_air)

        thread_tw.start()
        thread_air.start()

        thread_tw.join()
        thread_air.join()
    else:
        run_tw(in_css_path, out_css_path, False)
        run_air()


if __name__ == "__main__":
    main()
