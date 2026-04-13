import Chart from 'chart.js/auto';

let images = [];
let current = 0;

document.addEventListener('DOMContentLoaded', function () {
    // ── Lightbox ──────────────────────────────────────────────────────────────
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
            if (e.target === lightbox)
                lightbox.classList.replace('flex', 'hidden');
        });
    }
    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const closeBtn = document.getElementById('close-btn');
    if (prevBtn)
        prevBtn.addEventListener('click', function (e) {
            e.stopPropagation();
            current = (current - 1 + images.length) % images.length;
            update();
        });
    if (nextBtn)
        nextBtn.addEventListener('click', function (e) {
            e.stopPropagation();
            current = (current + 1) % images.length;
            update();
        });
    if (closeBtn)
        closeBtn.addEventListener('click', function () {
            lightbox.classList.replace('flex', 'hidden');
        });

    // ── Tabs ──────────────────────────────────────────────────────────────────
    document.querySelectorAll('[data-tab]').forEach(function (btn) {
        btn.addEventListener('click', function () {
            switchTab(btn.dataset.tab);
        });
    });

    // ── Description expand/collapse ───────────────────────────────────────────
    const descWrapper = document.getElementById('desc-wrapper');
    const descToggle = document.getElementById('desc-toggle');
    const descFade = document.getElementById('desc-fade');
    if (descWrapper && descToggle) {
        if (descWrapper.scrollHeight <= descWrapper.offsetHeight + 2) {
            if (descFade) descFade.style.display = 'none';
            descToggle.style.display = 'none';
        } else {
            let expanded = false;
            descToggle.addEventListener('click', function () {
                expanded = !expanded;
                descWrapper.style.maxHeight = expanded
                    ? descWrapper.scrollHeight + 'px'
                    : '10rem';
                descToggle.textContent = expanded
                    ? descToggle.dataset.less
                    : descToggle.dataset.more;
                if (descFade) descFade.style.display = expanded ? 'none' : '';
            });
        }
    }

    // ── HTMX: init chart after history panel loads ────────────────────────────
    document.addEventListener('htmx:afterSwap', function (e) {
        if (e.detail.target && e.detail.target.id === 'panel-history') {
            switchTab('history');
            initPriceChart();
        }
    });
});

function update() {
    document.getElementById('lightbox-img').src = images[current];
    document.getElementById('lightbox-counter').textContent =
        current + 1 + ' / ' + images.length;
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

function switchTab(tab) {
    const allPanels = ['panel-desc', 'panel-specs', 'panel-history'];
    const allTabs = ['tab-desc', 'tab-specs', 'tab-history'];

    allPanels.forEach(function (id) {
        const el = document.getElementById(id);
        if (el) el.classList.add('hidden');
    });
    allTabs.forEach(function (id) {
        const el = document.getElementById(id);
        if (!el) return;
        el.classList.remove('text-blue-600', 'border-blue-600');
        el.classList.add('text-gray-500', 'border-transparent');
    });

    const activePanel = document.getElementById('panel-' + tab);
    const activeTab = document.getElementById('tab-' + tab);
    if (activePanel) activePanel.classList.remove('hidden');
    if (activeTab) {
        activeTab.classList.add('text-blue-600', 'border-blue-600');
        activeTab.classList.remove('text-gray-500', 'border-transparent');
    }
}

function initPriceChart() {
    const canvas = document.getElementById('price-history-chart');
    if (!canvas) return;

    const labels = JSON.parse(canvas.dataset.labels);
    const prices = JSON.parse(canvas.dataset.prices);
    const originals = JSON.parse(canvas.dataset.originals);
    const discounts = JSON.parse(canvas.dataset.discounts);

    // Compute stats
    const validPrices = prices.filter(function (p) {
        return p > 0;
    });
    if (validPrices.length > 0) {
        const low = Math.min.apply(null, validPrices);
        const high = Math.max.apply(null, validPrices);
        const avg =
            validPrices.reduce(function (a, b) {
                return a + b;
            }, 0) / validPrices.length;
        const fmtNum = function (n) {
            return new Intl.NumberFormat('vi-VN', {
                style: 'decimal',
                maximumFractionDigits: 0,
            }).format(n);
        };
        const statLow = document.getElementById('stat-low');
        const statHigh = document.getElementById('stat-high');
        const statAvg = document.getElementById('stat-avg');
        if (statLow) statLow.textContent = fmtNum(low);
        if (statHigh) statHigh.textContent = fmtNum(high);
        if (statAvg) statAvg.textContent = fmtNum(avg);
    }

    new Chart(canvas, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Price',
                    data: prices,
                    borderColor: '#3b82f6',
                    backgroundColor: 'rgba(59,130,246,0.08)',
                    borderWidth: 2,
                    pointRadius: 3,
                    pointHoverRadius: 5,
                    fill: true,
                    tension: 0.3,
                    yAxisID: 'yPrice',
                },
                {
                    label: 'Original Price',
                    data: originals,
                    borderColor: '#d1d5db',
                    borderDash: [5, 4],
                    borderWidth: 1.5,
                    pointRadius: 0,
                    fill: false,
                    tension: 0.3,
                    yAxisID: 'yPrice',
                },
                {
                    label: 'Discount %',
                    data: discounts,
                    borderColor: '#fb923c',
                    backgroundColor: 'rgba(251,146,60,0.06)',
                    borderWidth: 1.5,
                    pointRadius: 2,
                    fill: false,
                    tension: 0.3,
                    yAxisID: 'yDiscount',
                },
            ],
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: { mode: 'index', intersect: false },
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: function (ctx) {
                            const val = ctx.parsed.y;
                            if (ctx.dataset.yAxisID === 'yDiscount') {
                                return ctx.dataset.label + ': ' + val + '%';
                            }
                            return (
                                ctx.dataset.label +
                                ': ' +
                                new Intl.NumberFormat('vi-VN').format(val)
                            );
                        },
                    },
                },
            },
            scales: {
                x: {
                    ticks: {
                        maxTicksLimit: 8,
                        font: { size: 11 },
                        color: '#9ca3af',
                    },
                    grid: { display: false },
                },
                yPrice: {
                    position: 'left',
                    ticks: {
                        font: { size: 11 },
                        color: '#9ca3af',
                        callback: function (v) {
                            if (v >= 1000000)
                                return (v / 1000000).toFixed(1) + 'M';
                            if (v >= 1000) return (v / 1000).toFixed(0) + 'K';
                            return v;
                        },
                    },
                    grid: { color: 'rgba(0,0,0,0.04)' },
                },
                yDiscount: {
                    position: 'right',
                    min: 0,
                    max: 100,
                    ticks: {
                        font: { size: 11 },
                        color: '#fb923c',
                        callback: function (v) {
                            return v + '%';
                        },
                    },
                    grid: { display: false },
                },
            },
        },
    });
}
