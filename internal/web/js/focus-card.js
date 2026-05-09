const grid = document.getElementById('results-grid');
const overlay = document.getElementById('overlay');
const focusedContent = document.getElementById('focused-content');
const closeBtn = document.getElementById('close-btn');
const nonglow_cards = document.querySelectorAll('.ip-card');
const glow_cards = document.querySelectorAll('.ip-card-glow');
let isDragging = false;
let startX = 0;
let startY = 0;

// Open focus card (only if user actually clicked, avoids opening on drag click)
function applyCardListeners(cards, cardClass) {
    cards.forEach((card) => {
        card.addEventListener('mousedown', (e) => {
            isDragging = false;
            startX = e.clientX;
            startY = e.clientY;
        });
        card.addEventListener('mousemove', (e) => {
            const dx = Math.abs(e.clientX - startX);
            const dy = Math.abs(e.clientY - startY);
            if (dx > 5 || dy > 5) {
                isDragging = true;
            }
        });
        card.addEventListener('click', () => {
            if (isDragging) return;
            focusedContent.innerHTML = '';
            const clone = card.cloneNode(true);
            clone.querySelectorAll('.info-hidden').forEach(el => {
                el.classList.remove('info-hidden');
                el.classList.add('info-row');
            });
            clone.classList.remove(cardClass);
            focusedContent.appendChild(clone);
            overlay.classList.add('active');
        });
    });
}

applyCardListeners(nonglow_cards, 'ip-card');
applyCardListeners(glow_cards, 'ip-card-glow');

// Close logic
closeBtn.addEventListener('click', () => {
    overlay.classList.remove('active');
});

overlay.addEventListener('click', (e) => {
if (e.target === overlay) {
    overlay.classList.remove('active');
}
});