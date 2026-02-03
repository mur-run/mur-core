-- murmur.nvim command implementations

local M = {}

local config = {
  mur_path = 'mur',
  notify_style = 'notify',
  use_telescope = true,
}

--- Set configuration from init.lua
function M.set_config(cfg)
  config = cfg
end

--- Send notification to user
---@param msg string
---@param level number|nil vim.log.levels
local function notify(msg, level)
  level = level or vim.log.levels.INFO
  if config.notify_style == 'notify' then
    vim.notify(msg, level, { title = 'Murmur' })
  else
    print('[Murmur] ' .. msg)
  end
end

--- Run mur command asynchronously
---@param args table Command arguments
---@param opts table|nil Options: on_stdout, on_exit, capture
---@return number Job ID
local function run_mur(args, opts)
  opts = opts or {}
  local cmd = vim.list_extend({ config.mur_path }, args)
  local output = {}

  return vim.fn.jobstart(cmd, {
    stdout_buffered = true,
    stderr_buffered = true,
    on_stdout = function(_, data)
      if data and data[1] ~= '' then
        vim.list_extend(output, data)
        if opts.on_stdout then
          opts.on_stdout(data)
        end
      end
    end,
    on_stderr = function(_, data)
      if data and data[1] ~= '' then
        for _, line in ipairs(data) do
          if line ~= '' then
            notify(line, vim.log.levels.WARN)
          end
        end
      end
    end,
    on_exit = function(_, code)
      if opts.on_exit then
        opts.on_exit(code, output)
      end
    end,
  })
end

--- Create a floating window with content
---@param lines table Lines to display
---@param title string Window title
local function show_float(lines, title)
  local width = 60
  local height = math.min(#lines + 2, 20)

  -- Calculate dimensions
  for _, line in ipairs(lines) do
    width = math.max(width, #line + 4)
  end
  width = math.min(width, vim.o.columns - 4)

  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)

  -- Create buffer
  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', false)
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')

  -- Create window
  local win = vim.api.nvim_open_win(buf, true, {
    relative = 'editor',
    width = width,
    height = height,
    row = row,
    col = col,
    style = 'minimal',
    border = 'rounded',
    title = ' ' .. title .. ' ',
    title_pos = 'center',
  })

  -- Keymaps to close
  vim.keymap.set('n', 'q', function()
    vim.api.nvim_win_close(win, true)
  end, { buffer = buf })
  vim.keymap.set('n', '<Esc>', function()
    vim.api.nvim_win_close(win, true)
  end, { buffer = buf })

  return buf, win
end

--- Sync all murmur sources
function M.sync()
  notify('Syncing all sources...')
  run_mur({ 'sync', 'all' }, {
    on_exit = function(code)
      if code == 0 then
        notify('Sync complete ✓')
      else
        notify('Sync failed (exit code: ' .. code .. ')', vim.log.levels.ERROR)
      end
    end,
  })
end

--- Extract patterns
function M.extract()
  notify('Extracting patterns...')
  run_mur({ 'learn', 'extract', '--auto' }, {
    on_exit = function(code, output)
      if code == 0 then
        local count = 0
        for _, line in ipairs(output) do
          if line:match('extracted') or line:match('pattern') then
            count = count + 1
          end
        end
        notify('Extraction complete ✓')
      else
        notify('Extraction failed (exit code: ' .. code .. ')', vim.log.levels.ERROR)
      end
    end,
  })
end

--- Show statistics in a floating window
function M.stats()
  notify('Fetching statistics...')
  run_mur({ 'stats' }, {
    on_exit = function(code, output)
      if code == 0 and #output > 0 then
        show_float(output, 'Murmur Statistics')
      elseif code == 0 then
        notify('No statistics available')
      else
        notify('Failed to fetch statistics', vim.log.levels.ERROR)
      end
    end,
  })
end

--- Browse patterns with Telescope or fallback to split
function M.patterns()
  -- Check if Telescope is available and enabled
  local has_telescope, telescope = pcall(require, 'telescope')
  local has_pickers = pcall(require, 'telescope.pickers')

  if config.use_telescope and has_telescope and has_pickers then
    M.patterns_telescope()
  else
    M.patterns_split()
  end
end

--- Show patterns in a split window (fallback)
function M.patterns_split()
  notify('Loading patterns...')
  run_mur({ 'learn', 'list' }, {
    on_exit = function(code, output)
      if code == 0 and #output > 0 then
        -- Create a new split
        vim.cmd('botright new')
        local buf = vim.api.nvim_get_current_buf()
        vim.api.nvim_buf_set_name(buf, 'murmur://patterns')
        vim.api.nvim_buf_set_lines(buf, 0, -1, false, output)
        vim.api.nvim_buf_set_option(buf, 'modifiable', false)
        vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
        vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
        vim.api.nvim_buf_set_option(buf, 'filetype', 'murmur-patterns')

        -- Close with q
        vim.keymap.set('n', 'q', ':bdelete<CR>', { buffer = buf, silent = true })
      elseif code == 0 then
        notify('No patterns found')
      else
        notify('Failed to load patterns', vim.log.levels.ERROR)
      end
    end,
  })
end

--- Show patterns with Telescope picker
function M.patterns_telescope()
  local pickers = require('telescope.pickers')
  local finders = require('telescope.finders')
  local conf = require('telescope.config').values
  local actions = require('telescope.actions')
  local action_state = require('telescope.actions.state')

  -- First, get the patterns
  run_mur({ 'learn', 'list', '--json' }, {
    on_exit = function(code, output)
      if code ~= 0 then
        notify('Failed to load patterns', vim.log.levels.ERROR)
        return
      end

      -- Try to parse JSON, fallback to raw lines
      local patterns = {}
      local raw_output = table.concat(output, '\n')
      local ok, parsed = pcall(vim.json.decode, raw_output)

      if ok and type(parsed) == 'table' then
        for _, p in ipairs(parsed) do
          table.insert(patterns, {
            display = p.name or p.pattern or tostring(p),
            value = p,
          })
        end
      else
        -- Fallback: use raw lines
        for _, line in ipairs(output) do
          if line ~= '' then
            table.insert(patterns, {
              display = line,
              value = { name = line },
            })
          end
        end
      end

      if #patterns == 0 then
        notify('No patterns found')
        return
      end

      -- Schedule Telescope picker on main thread
      vim.schedule(function()
        pickers.new({}, {
          prompt_title = 'Murmur Patterns',
          finder = finders.new_table({
            results = patterns,
            entry_maker = function(entry)
              return {
                value = entry.value,
                display = entry.display,
                ordinal = entry.display,
              }
            end,
          }),
          sorter = conf.generic_sorter({}),
          attach_mappings = function(prompt_bufnr)
            actions.select_default:replace(function()
              actions.close(prompt_bufnr)
              local selection = action_state.get_selected_entry()
              if selection then
                notify('Selected: ' .. selection.display)
              end
            end)
            return true
          end,
        }):find()
      end)
    end,
  })
end

--- Health check
function M.health()
  local health = vim.health or require('health')
  local start = health.start or health.report_start
  local ok = health.ok or health.report_ok
  local warn = health.warn or health.report_warn
  local error_fn = health.error or health.report_error

  start('murmur.nvim')

  -- Check mur CLI
  local mur_version = vim.fn.system(config.mur_path .. ' --version')
  if vim.v.shell_error == 0 then
    ok('mur CLI found: ' .. mur_version:gsub('\n', ''))
  else
    error_fn('mur CLI not found. Install from: https://github.com/poyenc/murmur-ai')
  end

  -- Check Telescope
  local has_telescope = pcall(require, 'telescope')
  if has_telescope then
    ok('telescope.nvim available')
  else
    warn('telescope.nvim not found (optional, for :MurmurPatterns picker)')
  end
end

return M
