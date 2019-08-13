$(document).ready(function() {
  var setSignupButton2Sensitivity = function(validPassword) {
    var validDate = validateDate($('#datepicker').val());
    $('#signup-button-2').attr('disabled', !validPassword || !validDate);
  }

  $('#datepicker').datepicker({
    language: 'en',
    view: 'years',
    // onSelect: function(_formattedDate, date, _dp) {
    //   selectedDate = date;
    //   setSignupButton2Sensitivity(validatePassword($('#password')));
    // },
  });
  $('#datepicker').on('input', function() {
    setSignupButton2Sensitivity(validatePassword($('#password')));
  });

  var loggingIn = false;
  var signingUp = false;

  $('#email').on('input', function() {
    var validEmail = validateEmail($('#email').val());
    $('#login-button-1').attr('disabled', !validEmail);
    $('#signup-button-1').attr('disabled', !validEmail);
  })

  $('#password').on('input', function() {
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
    var email = $('#email').val();
    var password = $('#password').val();
    $.post('/login', {email, password}, function(_data, status) {
      console.log(`xxx ${status}`);
    });
  });

  $('#cancel-button').click(function() {
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

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
function validateEmail(mail) {
  return /^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/.test(mail);
}

function validatePassword(pw) {
  return !!pw;
}

function validateDate(date) {
  return true; // xxx
}
