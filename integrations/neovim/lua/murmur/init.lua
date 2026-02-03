-- murmur.nvim - Neovim integration for murmur-ai
-- https://github.com/poyenc/murmur-ai

local M = {}
local commands = require('murmur.commands')

-- Configuration with defaults
M.config = {
  -- Path to mur CLI (default: assumes it's in PATH)
  mur_path = 'mur',
  -- Notification style: 'notify' or 'print'
  notify_style = 'notify',
  -- Enable Telescope integration if available
  use_telescope = true,
}

--- Setup the plugin with user configuration
---@param opts table|nil User configuration
function M.setup(opts)
  M.config = vim.tbl_deep_extend('force', M.config, opts or {})
  commands.set_config(M.config)
end

--- Sync all murmur sources
function M.sync()
  commands.sync()
end

--- Extract patterns from current buffer or auto
function M.extract()
  commands.extract()
end

--- Show murmur statistics
function M.stats()
  commands.stats()
end

--- Browse patterns (Telescope if available, else split)
function M.patterns()
  commands.patterns()
end

--- Check if mur CLI is available
function M.health()
  commands.health()
end

return M
