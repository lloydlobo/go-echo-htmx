# notes

## tools

### neovim

Add the binaries "nvim.exe" and "nvim-qt.exe" to your Windows PATH

```shell
setx PATH "%PATH%;C:\Program Files\Neovim\bin"
```

#### Tips

##### Paste text

- Open the register containing the copied text. By default, this is done with the `"+p` command. The + register holds the most recently copied text.

##### Copy text

- Use the `"+y` command to copy the text to the named register "+".

##### Tabs

- Open files in tabs: `nvim -p file1.txt file2.txt` or use `:tabnew`, `:tabedit filename`, `:tabopen filename`
- Switch tabs: `gt`, `gT`, `:tabn` (or mouse if plugins enabled)

### templ

```shell
templ generate
```

```shell
templ fmt .
```

### air

```shell
air init
air
```
