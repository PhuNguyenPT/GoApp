let images = [];
let current = 0;

document.addEventListener('DOMContentLoaded', function () {
    const el = document.getElementById('lightbox-trigger');
    const lightbox = document.getElementById('lightbox');

    if (el) {
        images = JSON.parse(el.dataset.images);
        el.addEventListener('click', function () {
            current = 0;
            update();
            lightbox.classList.replace('hidden', 'flex');
        });
    }

    if (lightbox) {
        lightbox.addEventListener('click', function (e) {
            if (e.target === lightbox) {
                lightbox.classList.replace('flex', 'hidden');
            }
        });
    }

    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const closeBtn = document.getElementById('close-btn');

    if (prevBtn) prevBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        current = (current - 1 + images.length) % images.length;
        update();
    });

    if (nextBtn) nextBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        current = (current + 1) % images.length;
        update();
    });

    if (closeBtn) closeBtn.addEventListener('click', function () {
        lightbox.classList.replace('flex', 'hidden');
    });
});

function update() {
    document.getElementById('lightbox-img').src = images[current];
    document.getElementById('lightbox-counter').textContent = (current + 1) + ' / ' + images.length;
}

document.addEventListener('keydown', function (e) {
    const lightbox = document.getElementById('lightbox');
    if (!lightbox || lightbox.classList.contains('hidden')) return;
    if (e.key === 'ArrowLeft') {
        current = (current - 1 + images.length) % images.length;
        update();
    }
    if (e.key === 'ArrowRight') {
        current = (current + 1) % images.length;
        update();
    }
    if (e.key === 'Escape') lightbox.classList.replace('flex', 'hidden');
});