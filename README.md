
# greg

**greg** is a lightweight, terminal-based launcher and filter tool written in Go. It can operate in **dmenu mode** for filtering piped input or **apps mode** for launching `.desktop` applications.

---

## Features

* dmenu-style fuzzy search interface for terminal input.
* Search and launch GUI applications from `.desktop` files.
* Configurable colors, maximum visible items, and logging.
* Supports auto-detecting the terminal height if `max_items = -1`.
* Fully detached application launches (apps mode) so the launcher can exit immediately.

---

## Installation

Make sure you have Go installed, then run:

```bash
go install github.com/chriso345/greg@latest
```

Alternatively, clone the repo and build:

```bash
git clone https://github.com/chriso345/greg.git
cd greg
go build -o greg
```

---

## Usage

### apps Mode (launch GUI applications, default)

```bash
greg -m apps
```

* Lists all `.desktop` applications in `~/.local/share/applications`.
* Type to search and navigate with ↑/↓.
* Press **Enter** to launch the selected app.
* Applications are launched fully detached from the terminal.

### dmenu Mode (filter piped input)

```bash
ls /usr/bin | greg -m dmenu
```

* Type to filter the list of items.
* Navigate with ↑/↓ keys.
* Press **Enter** to select; the selected item is printed to stdout.

### menu Mode (select from a predefined multi-level menu)

```bash
greg -m menu
```

* Navigate a predefined multi-level menu structure defined in `~/.config/greg/menu.toml`.
* Type to filter menu items.
* Press **Enter** to select an item.
* Support an additional `--start/-s` flag to specify the starting menu id.

---

## Configuration

`greg` supports a TOML configuration file at:

```
$XDG_CONFIG_HOME/greg/config.toml
```

or

```
~/.config/greg/config.toml
```

### Example `~/.config/greg/config.toml`

```toml
# Number of items to show. Set -1 to auto-detect terminal height.
max_items = 8

# Enable logging for debugging
log = false

[colors]
title = "214"     # orange
prompt = "45"     # blue
item = "252"      # gray
selected = "54"   # teal background
help = "240"      # dim gray
```

* `max_items`: Maximum visible items in the TUI. `-1` auto-detects terminal height.
* `log`: Enables debug logging.
* `colors`: Terminal color codes for TUI elements.

---

## License

Licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
