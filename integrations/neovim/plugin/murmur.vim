" murmur.vim - Neovim plugin for murmur-ai
" Maintainer: murmur-ai team
" License: MIT

if exists('g:loaded_murmur') || &cp
  finish
endif
let g:loaded_murmur = 1

" Require Neovim 0.7+
if !has('nvim-0.7')
  echohl WarningMsg
  echom 'murmur.nvim requires Neovim 0.7 or later'
  echohl None
  finish
endif

" Commands
command! MurmurSync lua require('murmur').sync()
command! MurmurExtract lua require('murmur').extract()
command! MurmurStats lua require('murmur').stats()
command! MurmurPatterns lua require('murmur').patterns()

" Health check (for :checkhealth murmur)
function! health#murmur#check() abort
  lua require('murmur.commands').health()
endfunction
