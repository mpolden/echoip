/*jslint es5: true, indent: 2, browser: true*/

(function () {
  'use strict';
  var onLoad = function (event) {
    var select = document.querySelector('#select-command');
    select.addEventListener('change', function () {
      [].forEach.call(document.querySelectorAll('.command'), function (el) {
        el.innerHTML = this.value;
      }, this);
    });
  }
  document.addEventListener('DOMContentLoaded', onLoad);
})();
