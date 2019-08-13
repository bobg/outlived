$(document).ready(function() {
  var selectedDate;

  var setSignupButton2Sensitivity = function(validPassword) {
    $('#signup-button-2').attr('disabled', !validPassword || !selectedDate);
  }

  $('#datepicker').datepicker({
    view: 'years',
    onSelect: function(_formattedDate, date, _dp) {
      selectedDate = date;
      setSignupButton2Sensitivity(validatePassword($('#password')));
    },
  });

  var loggingIn = false;
  var signingUp = false;

  $('#email').change(function() {
    var validEmail = validateEmail($('#email').val());
    $('#login-button-1').attr('disabled', !validEmail);
    $('#signup-button-1').attr('disabled', !validEmail);
  })

  $('#password').change(function() {
    var validPassword = validatePassword($('#password'));
    $('#login-button-2').attr('disabled', !validPassword);
    setSignupButton2Sensitivity(validPassword);
  });

  var loggingInSigningUp = function() {
    $('#login-button-1').hide();
    $('#signup-button-1').hide();
    $('#password').show();
    $('#cancel-button').show();

    $('#email-value').text($('#email').val());
    $('#email-value').show();
    $('#email').hide()
  }

  $('#signup-button-1').click(function() {
    loggingInSigningUp();
    $('#datepicker').show();
    $('#signup-button-2').show();
  });

  $('#login-button-1').click(function() {
    loggingInSigningUp();
    $('#login-button-2').show();
  });

  $('#signup-button-2').click(function() {
    // xxx post to /signup: email, password, born
  });

  $('#login-button-2').click(function() {
    // xxx post to /login: email, password
  });

  $('#cancel').click(function() {
    $('#email').val('');
    $('#email').show();
    $('#email-value').hide();
    $('#signup-button-1').show();
    $('#login-button-1').show();
    $('#password').hide();
    $('#login-button-2').hide();
    $('#datepicker').hide();
    $('#signup-button-2').hide();
    $('#cancel-button').hide();
  });
});
