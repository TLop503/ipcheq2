const grid = document.getElementById('results-grid');
const overlay = document.getElementById('overlay');
const focusedContent = document.getElementById('focused-content');
const closeBtn = document.getElementById('close-btn');
const cards = document.querySelectorAll('.ip-card');

// Open focus card
cards.forEach((card, index) => {
  card.addEventListener('click', () => {
    overlay.classList.add('active');
  });
});

// Close logic
closeBtn.addEventListener('click', () => {
    overlay.classList.remove('active');
});

overlay.addEventListener('click', (e) => {
if (e.target === overlay) {
    overlay.classList.remove('active');
}
});