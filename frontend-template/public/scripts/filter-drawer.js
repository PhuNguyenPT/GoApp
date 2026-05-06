document.addEventListener('DOMContentLoaded', function () {
    const openBtn = document.getElementById('open-filter-drawer');
    const closeBtn = document.getElementById('close-filter-drawer');
    const drawer = document.getElementById('filter-drawer');
    const overlay = document.getElementById('filter-overlay');

    function openDrawer() {
        drawer.classList.remove('-translate-x-full');
        overlay.classList.remove('hidden');
        document.body.classList.add('overflow-hidden');
    }

    function closeDrawer() {
        drawer.classList.add('-translate-x-full');
        overlay.classList.add('hidden');
        document.body.classList.remove('overflow-hidden');
    }

    if (openBtn) openBtn.addEventListener('click', openDrawer);
    if (closeBtn) closeBtn.addEventListener('click', closeDrawer);
    if (overlay) overlay.addEventListener('click', closeDrawer);

    function scrollActiveFilterIntoView() {
        setTimeout(function () {
            document
                .querySelectorAll('#category-filter ul, #subcategory-filter ul')
                .forEach(function (list) {
                    const active = list.querySelector('[class*="bg-blue"]');
                    if (active)
                        list.scrollTop = active.offsetTop - list.offsetTop;
                });
        }, 50);
    }

    scrollActiveFilterIntoView();
    document.addEventListener('htmx:afterSettle', scrollActiveFilterIntoView);

    document.addEventListener('htmx:afterSettle', function () {
        document.querySelectorAll('img').forEach(function (img) {
            img.addEventListener('error', function () {
                this.style.display = 'none';
            });
        });
    });
});
