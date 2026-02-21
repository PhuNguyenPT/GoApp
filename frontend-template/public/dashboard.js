document.addEventListener('DOMContentLoaded', function () {
    const btn = document.getElementById('email-toggle');
    if (!btn) return;
    btn.addEventListener('click', function () {
        const display = document.getElementById('email-display');
        const showing = btn.textContent.trim() === 'Hide';
        display.textContent = showing ? btn.dataset.masked : btn.dataset.full;
        btn.textContent = showing ? 'Show' : 'Hide';
    });
});