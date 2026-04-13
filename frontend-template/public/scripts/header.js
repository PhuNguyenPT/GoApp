document.addEventListener('DOMContentLoaded', function () {
    const btn = document.getElementById('mobile-menu-btn');
    if (!btn) return;
    btn.addEventListener('click', function () {
        const target = document.getElementById(btn.dataset.target);
        if (target) target.classList.toggle('hidden');
    });
});
