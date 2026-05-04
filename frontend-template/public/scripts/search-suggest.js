document.addEventListener('DOMContentLoaded', function () {
    const input = document.getElementById('search-input');
    const drop = document.getElementById('search-dropdown');
    if (!input || !drop) return;

    let fetchTimer;

    function getItems() {
        return Array.from(drop.querySelectorAll('[data-value]'));
    }

    function getActiveIdx(items) {
        return items.findIndex(function (el) {
            return el.classList.contains('bg-blue-100');
        });
    }

    function setActive(items, idx) {
        items.forEach(function (el) {
            el.classList.remove('bg-blue-100');
        });
        if (idx >= 0 && idx < items.length) {
            items[idx].classList.add('bg-blue-100');
        }
    }

    function selectItem(value) {
        input.value = value;
        drop.classList.add('hidden');
        drop.innerHTML = '';
        // Use htmx.trigger so HTMX owns the request — hx-push-url then works correctly.
        htmx.trigger(input, 'search');
    }

    input.addEventListener('input', function () {
        const q = input.value;
        clearTimeout(fetchTimer);
        if (q.length < 2) {
            drop.classList.add('hidden');
            return;
        }
        fetchTimer = setTimeout(function () {
            fetch('/products/suggest?q=' + encodeURIComponent(q))
                .then(function (r) {
                    return r.text();
                })
                .then(function (html) {
                    drop.innerHTML = html;
                    drop.classList.toggle('hidden', !html.trim());
                });
        }, 300);
    });

    input.addEventListener('keydown', function (e) {
        // Always allow Escape, even when dropdown is hidden.
        if (e.key === 'Escape') {
            e.preventDefault(); // stop browser clearing the search input
            drop.classList.add('hidden');
            return;
        }

        // Arrow / Enter navigation only makes sense when the dropdown is visible.
        if (drop.classList.contains('hidden')) return;
        const items = getItems();
        if (!items.length) return;

        // Prevent arrow keys from restarting the fetch timer via the 'input' listener.
        // (Arrow keys don't fire 'input', but guard here for safety.)

        let idx = getActiveIdx(items);

        if (e.key === 'ArrowDown') {
            e.preventDefault();
            // If nothing selected (-1), go to first item.
            idx = (idx + 1) % items.length;
            setActive(items, idx);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            // If nothing selected (-1), go to last item (not second-to-last).
            idx = idx <= 0 ? items.length - 1 : idx - 1;
            setActive(items, idx);
        } else if (e.key === 'Enter') {
            const active = items[idx];
            if (active) {
                e.preventDefault();
                selectItem(active.dataset.value);
            }
        }
    });

    input.addEventListener('blur', function () {
        setTimeout(function () {
            drop.classList.add('hidden');
        }, 150);
    });

    input.addEventListener('focus', function () {
        if (input.value.length >= 2 && drop.innerHTML.trim()) {
            drop.classList.remove('hidden');
        }
    });

    // mousedown fires before blur; e.preventDefault() keeps focus on the input
    // so the blur handler doesn't hide the dropdown before selectItem runs.
    drop.addEventListener('mousedown', function (e) {
        const item = e.target.closest('[data-value]');
        if (!item) return;
        e.preventDefault();
        selectItem(item.dataset.value);
    });
});
