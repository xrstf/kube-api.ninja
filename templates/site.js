document.querySelector('body').classList.add('has-javascript');

// hide accordion items on the about page
document.querySelectorAll('.accordion-button').forEach(function(n) {
  n.classList.add('collapsed');
});

document.querySelectorAll('.accordion-collapse').forEach(function(n) {
  n.classList.add('collapse');
});
