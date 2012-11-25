$(document).ready ->
  $('#select-command').on 'change', ->
    $('.command').text $(this).val()
