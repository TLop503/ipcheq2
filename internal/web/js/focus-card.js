const grid = document.getElementById('results-grid');
const overlay = document.getElementById('overlay');
const focusedContent = document.getElementById('focused-content');
const closeBtn = document.getElementById('close-btn');
const cards = document.querySelectorAll('.ip-card');
let isDragging = false;
let startX = 0;
let startY = 0;

// Open focus card (only if user actually clicked, avoids opening on drag click)
cards.forEach((card) => {
  card.addEventListener('mousedown', (e) => {
    isDragging = false;
    startX = e.clientX;
    startY = e.clientY;
  });

  card.addEventListener('mousemove', (e) => {
    const dx = Math.abs(e.clientX - startX);
    const dy = Math.abs(e.clientY - startY);

    if (dx > 5 || dy > 5) { // small threshold
      isDragging = true;
    }
  });

  card.addEventListener('click', () => {
    if (isDragging) return; // ignore drag

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