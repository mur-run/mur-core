# murmur.nvim

Neovim plugin for [murmur-ai](https://github.com/mur-run/mur-core) — sync, extract, and browse your learning patterns without leaving the editor.

## Requirements

- Neovim 0.7+
- `mur` CLI installed and in PATH
- Optional: [telescope.nvim](https://github.com/nvim-telescope/telescope.nvim) for enhanced pattern browsing

## Installation

### [lazy.nvim](https://github.com/folke/lazy.nvim)

```lua
{
  'mur-run/mur-core',
  config = function()
    require('murmur').setup({
      -- optional configuration
      mur_path = 'mur',  -- path to mur CLI
      use_telescope = true,  -- use telescope for pattern picker
    })
  end,
  -- If installed as a subdirectory plugin:
  -- dir = '~/Projects/murmur-ai/integrations/neovim',
}
```

### [packer.nvim](https://github.com/wbthomason/packer.nvim)

```lua
use {
  'mur-run/mur-core',
  config = function()
    require('murmur').setup()
  end
}

-- Or from local path:
use {
  '~/Projects/murmur-ai/integrations/neovim',
  config = function()
    require('murmur').setup()
  end
}
```

### [vim-plug](https://github.com/junegunn/vim-plug)

```vim
" From GitHub (if published as standalone repo)
Plug 'poyenc/murmur.nvim'

" Or from local path:
Plug '~/Projects/murmur-ai/integrations/neovim'

" After plug#end(), add:
lua require('murmur').setup()
```

### Manual Installation

```bash
# Clone the repo or symlink
mkdir -p ~/.local/share/nvim/site/pack/murmur/start
ln -s ~/Projects/murmur-ai/integrations/neovim ~/.local/share/nvim/site/pack/murmur/start/murmur.nvim
```

Then add to your `init.lua`:

```lua
require('murmur').setup()
```

## Commands

| Command | Description |
|---------|-------------|
| `:MurmurSync` | Sync all configured sources |
| `:MurmurExtract` | Extract patterns automatically |
| `:MurmurStats` | Show statistics in a floating window |
| `:MurmurPatterns` | Browse patterns (Telescope picker or split) |

## Configuration

```lua
require('murmur').setup({
  -- Path to mur CLI (default: 'mur', assumes it's in PATH)
  mur_path = 'mur',

  -- Notification style: 'notify' (vim.notify) or 'print'
  notify_style = 'notify',

  -- Use Telescope for pattern browsing if available
  use_telescope = true,
})
```

## Keymaps

The plugin doesn't set any keymaps by default. Here are some suggested mappings:

```lua
-- In your init.lua or keymaps file
vim.keymap.set('n', '<leader>ms', '<cmd>MurmurSync<cr>', { desc = 'Murmur: Sync all' })
vim.keymap.set('n', '<leader>me', '<cmd>MurmurExtract<cr>', { desc = 'Murmur: Extract patterns' })
vim.keymap.set('n', '<leader>mt', '<cmd>MurmurStats<cr>', { desc = 'Murmur: Show stats' })
vim.keymap.set('n', '<leader>mp', '<cmd>MurmurPatterns<cr>', { desc = 'Murmur: Browse patterns' })
```

## Health Check

Run `:checkhealth murmur` to verify your setup.

## License

MIT
