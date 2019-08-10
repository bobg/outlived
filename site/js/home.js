$(document).ready(function() {
  $(#datepicker).datepicker({view: 'years'});
  $(#datepicker).hide();

  $(#signup-button).click(function() {
    $(#datepicker).show();
  });

  $(#email).change(function() {
    setButtonSensitivity();
  });
  $(#password).change(function() {
    setButtonSensitivity();
  });
});

function setButtonSensitivity() {
  var validEmail = validateEmail($(#email).val());
  var validPassword = validatePassword($(#password).val());
  $(#signup-button).attr('disabled', !validEmail || !validPassword);
  $(#login-button).attr('disabled', !validEmail || !validPassword);
}
