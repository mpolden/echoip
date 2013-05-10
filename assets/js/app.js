/*jslint es5: true, indent: 2, browser: true*/
/*global jQuery: true*/

(function ($) {
  'use strict';

  $(document).ready(function () {
   $('#select-command').on('change', function () {
    $('.command').text($(this).val());
   });
  });

})(jQuery);

