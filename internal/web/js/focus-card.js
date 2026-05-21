const grid = document.getElementById('results-grid');
const overlay = document.getElementById('overlay');
const focused_content = document.getElementById('focused-content');
const close_btn = document.getElementById('close-btn');
const nonglow_cards = document.querySelectorAll('.ip-card');
const glow_cards = document.querySelectorAll('.ip-card-glow');
const focused_card = document.getElementById('focused');
let is_dragging = false;
let startX = 0;
let startY = 0;

// Open focus card (only if user actually clicked, avoids opening on drag click)
function applyCardListeners(cards, cardClass) {
    cards.forEach((card) => {
        card.addEventListener('mousedown', (e) => {
            is_dragging = false;
            startX = e.clientX;
            startY = e.clientY;
        });
        card.addEventListener('mousemove', (e) => {
            const dx = Math.abs(e.clientX - startX);
            const dy = Math.abs(e.clientY - startY);
            if (dx > 5 || dy > 5) {
                is_dragging = true;
            }
        });
        card.addEventListener('click', () => {
            if (is_dragging) return;
            focused_content.innerHTML = '';
            const clone = card.cloneNode(true);
            const confidence_bar = clone.getElementsByClassName('confidence-bar');
            confidence_bar[0].remove();
            clone.querySelectorAll('.info-hidden').forEach(el => {
                el.classList.remove('info-hidden');
                el.classList.add('info-row');
            });
            clone.classList.remove(cardClass);
            focused_content.appendChild(clone);
            overlay.classList.add('active');
            if (cardClass == 'ip-card-glow') {
                // make the focus card glow too
                focused_card.classList.add('glow');
            }
            else {
                focused_card.classList.remove('glow');
            }
        });
    });
}

applyCardListeners(nonglow_cards, 'ip-card');
applyCardListeners(glow_cards, 'ip-card-glow');

// Close logic
close_btn.addEventListener('click', () => {
    overlay.classList.remove('active');
});

overlay.addEventListener('click', (e) => {
if (e.target === overlay) {
    overlay.classList.remove('active');
}
});