$(document).ready(function() {
  var setSignupButton2Sensitivity = function(validPassword) {
    var validDate = validateDate($('#datepicker').val());
    $('#signup-button-2').attr('disabled', !validPassword || !validDate);
  }

  $('#datepicker').datepicker({
    dateFormat: 'yyyy-mm-dd',
    language: 'en',
    view: 'years',
    onSelect: function() {
      setSignupButton2Sensitivity(validatePassword($('#password').val()));
    },
  });
  $('#datepicker').on('input', function() {
    setSignupButton2Sensitivity(validatePassword($('#password').val()));
  });

  var loggingIn = false;
  var signingUp = false;

  $('#email').on('input', function() {
    var validEmail = validateEmail($('#email').val());
    $('#login-button-1').attr('disabled', !validEmail);
    $('#signup-button-1').attr('disabled', !validEmail);
  })

  $('#password').on('input', function() {
    var validPassword = validatePassword($('#password').val());
    $('#login-button-2').attr('disabled', !validPassword);
    setSignupButton2Sensitivity(validPassword);
  });

  var loggingInSigningUp = function() {
    $('#login-button-1').hide();
    $('#signup-button-1').hide();
    $('#password-div').show();
    $('#cancel-button').show();

    $('#email-value').text($('#email').val());
    $('#email-value').show();
    $('#email').hide()
  }

  $('#signup-button-1').click(function() {
    loggingInSigningUp();
    $('#datepicker-div').show();
    $('#signup-button-2').show();
    $('#login-signup-form').attr('action', '/signup')
  });

  $('#login-button-1').click(function() {
    loggingInSigningUp();
    $('#login-button-2').show();
    $('#login-signup-form').attr('action', '/login')
  });

  $('#cancel-button').click(function() {
    $('#email').val('');
    $('#email').show();
    $('#email-value').hide();
    $('#signup-button-1').show();
    $('#login-button-1').show();
    $('#password-div').hide();
    $('#login-button-2').hide();
    $('#datepicker-div').hide();
    $('#signup-button-2').hide();
    $('#cancel-button').hide();
  });

  $('#active').click(function() {
    var wasChecked = $('#active').attr('checked');
    $('#active').attr('disabled', true);
    $.ajax({
      url: '/setactive',
      method: 'POST',
      data: {active: !wasChecked},
      success: () => {
        $('#active').attr('checked', !wasChecked);
      },
      error: () => {
        $('#active').attr('checked', wasChecked);
      },
      complete: () => {
        $('#active').attr('disabled', false);
      },
    });
  })
});

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
function validateEmail(mail) {
  return /^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/.test(mail);
}

function validatePassword(pw) {
  return !!pw;
}

var monthDays = [0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];

function validateDate(date) {
  var m = date.match(/(\d+)\D+(\d+)\D+(\d+)/);
  if (!m) { return false }

  var today = new Date();
  var y = parseInt(m[1]);
  if (y < 1900) { return false }
  if (y > today.getFullYear()) { return false }

  var mon = parseInt(m[2]);
  if (mon < 1 || mon > 12) { return false }

  var d = parseInt(m[3]);
  if (d < 1) { return false }

  if (d > monthDays[mon]) { return false }

  return true;
}
