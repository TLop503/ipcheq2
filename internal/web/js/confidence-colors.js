// Dynamic confidence bar and badge color based on abuse confidence percentage
document.addEventListener('DOMContentLoaded', function() {
    const cards = document.querySelectorAll('.ip-card');

    cards.forEach(card => {
        const confidence = parseInt(card.getAttribute('data-confidence'));
        const confidenceBar = card.querySelector('.confidence-bar');
        const confidenceBadge = card.querySelector('.confidence-badge');

        // Determine color range and gradient based on confidence level
        let gradient;
        let badgeColor;

        if (confidence === 0) {
            // 0% - Solid green
            gradient = '#22c55e';
            badgeColor = '#22c55e';
        } else if (confidence <= 25) {
            // 1-25% - Solid green
            gradient = '#22c55e';
            badgeColor = '#22c55e';
        } else if (confidence <= 50) {
            // 26-50% - Green to Yellow gradient
            gradient = 'linear-gradient(to right, #22c55e, #eab308)';
            const ratio = (confidence - 25) / 25;
            badgeColor = interpolateColor('#22c55e', '#eab308', ratio);
        } else if (confidence <= 75) {
            // 51-75% - Yellow to Orange gradient
            gradient = 'linear-gradient(to right, #eab308, #f97316)';
            const ratio = (confidence - 50) / 25;
            badgeColor = interpolateColor('#eab308', '#f97316', ratio);
        } else if (confidence < 100) {
            // 76-99% - Orange to Red gradient
            gradient = 'linear-gradient(to right, #f97316, #ef4444)';
            const ratio = (confidence - 75) / 25;
            badgeColor = interpolateColor('#f97316', '#ef4444', ratio);
        } else {
            // 100% - Solid red
            gradient = '#ef4444';
            badgeColor = '#ef4444';
        }

        // Apply colors
        if (confidenceBar) {
            confidenceBar.style.background = gradient;
        }

        if (confidenceBadge) {
            confidenceBadge.style.backgroundColor = badgeColor;
        }
    });
});

// Helper function to interpolate between two hex colors
function interpolateColor(color1, color2, ratio) {
    const hex = (color) => {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(color);
        return result ? {
            r: parseInt(result[1], 16),
            g: parseInt(result[2], 16),
            b: parseInt(result[3], 16)
        } : null;
    };

    const c1 = hex(color1);
    const c2 = hex(color2);

    if (!c1 || !c2) return color1;

    const r = Math.round(c1.r + (c2.r - c1.r) * ratio);
    const g = Math.round(c1.g + (c2.g - c1.g) * ratio);
    const b = Math.round(c1.b + (c2.b - c1.b) * ratio);

    return `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
}