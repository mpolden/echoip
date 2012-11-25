$(document).ready ->
  $('#select-command').bind 'change', ->
    $('.command').text $(this).val()
