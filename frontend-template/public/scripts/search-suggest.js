document.addEventListener('DOMContentLoaded', function () {
    const input = document.getElementById('search-input');
    const drop = document.getElementById('search-dropdown');
    if (!input || !drop) return;

    let t;

    input.addEventListener('input', function () {
        clearTimeout(t);
        const q = input.value;
        if (q.length < 2) {
            drop.classList.add('hidden');
            return;
        }
        t = setTimeout(function () {
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

    input.addEventListener('blur', function () {
        setTimeout(function () {
            drop.classList.add('hidden');
        }, 150);
    });

    input.addEventListener('focus', function () {
        if (input.value.length >= 2) drop.classList.remove('hidden');
    });
});
