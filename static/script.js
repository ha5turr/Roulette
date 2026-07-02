const canvas = document.getElementById('wheel');
const ctx = canvas.getContext('2d');
const resultOverlay = document.getElementById('result-overlay');
const resultName = document.getElementById('result-name');
const resultDesc = document.getElementById('result-desc');
const resultImage = document.getElementById('result-image');
const rouletteSelect = document.getElementById('roulette-select');

let currentRoulette = null;
let events = [];
let currentAngle = 0;
let spinning = false;

async function loadRoulettes() {
    const response = await fetch('/api/roulettes');
    const names = await response.json();
    rouletteSelect.innerHTML = '';
    names.forEach(name => {
        const option = document.createElement('option');
        option.value = name;
        option.textContent = name;
        rouletteSelect.appendChild(option);
    });
    if (names.length > 0) {
        rouletteSelect.value = names[0];
        loadRouletteEvents(names[0]);
    }
}

async function loadRouletteEvents(name) {
    const response = await fetch(`/api/roulette/${name}`);
    const data = await response.json();
    events = data;
    currentRoulette = name;
    currentAngle = 0;
    drawWheel(events);
    resultOverlay.style.display = 'none';
}

function drawWheel(events) {
    const n = events.length;
    if (n === 0) return;
    const angleStep = (2 * Math.PI) / n;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (let i = 0; i < n; i++) {
        const startAngle = i * angleStep;
        const endAngle = startAngle + angleStep;
        ctx.beginPath();
        ctx.moveTo(250, 250);
        ctx.arc(250, 250, 250, startAngle, endAngle);
        ctx.closePath();
        ctx.fillStyle = events[i].color || `hsl(${i * 360 / n}, 70%, 50%)`;
        ctx.fill();
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 2;
        ctx.stroke();

        const midAngle = startAngle + angleStep / 2;
        // Текст размещаем на радиусе 170 (чуть ближе к центру, чтобы не вылезал)
        const textRadius = 170;
        const textX = 250 + textRadius * Math.cos(midAngle);
        const textY = 250 + textRadius * Math.sin(midAngle);
        ctx.save();
        ctx.translate(textX, textY);
        // Поворачиваем так, чтобы текст был вдоль касательной (по дуге)
        // Для читаемости: если сектор в нижней половине, разворачиваем текст
        ctx.rotate(midAngle);
        ctx.fillStyle = '#fff';
        ctx.font = 'bold 14px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        // Ограничим длину текста, чтобы не вылезал
        let label = events[i].name;
        if (label.length > 12) label = label.slice(0, 10) + '…';
        ctx.fillText(label, 0, 0);
        ctx.restore();
    }
}

function getSectorAtTop() {
    const topAngle = 3 * Math.PI / 2;
    let angle = currentAngle % (2 * Math.PI);
    if (angle < 0) angle += 2 * Math.PI;
    let sectorAngle = (topAngle - angle + 2 * Math.PI) % (2 * Math.PI);
    const n = events.length;
    const angleStep = 2 * Math.PI / n;
    for (let i = 0; i < n; i++) {
        const start = i * angleStep;
        const end = (i + 1) * angleStep;
        if (sectorAngle >= start && sectorAngle < end) {
            return i;
        }
    }
    return -1;
}

function spinToEvent(targetEvent) {
    if (spinning) return;
    spinning = true;
    resultOverlay.style.display = 'none';

    const idx = events.findIndex(e => e.id === targetEvent.id);
    console.log('🎯 spinToEvent target:', targetEvent.id, 'idx:', idx);
    if (idx === -1) {
        spinning = false;
        return;
    }

    const n = events.length;
    const angleStep = (2 * Math.PI) / n;
    const targetAngle = idx * angleStep + angleStep / 2;
    const extraSpins = 8 + Math.floor(Math.random() * 5);
    let finalAngle = (3 * Math.PI / 2) - targetAngle + 2 * Math.PI * extraSpins;
    const minRotation = 6 * 2 * Math.PI;
    while (finalAngle <= currentAngle + minRotation) {
        finalAngle += 2 * Math.PI;
    }
    console.log('📐 currentAngle:', currentAngle, 'finalAngle:', finalAngle);

    const startAngle = currentAngle;
    const duration = 3000;
    const startTime = performance.now();

    function animate(now) {
        const elapsed = now - startTime;
        const progress = Math.min(elapsed / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3);
        currentAngle = startAngle + (finalAngle - startAngle) * eased;

        ctx.save();
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.translate(250, 250);
        ctx.rotate(currentAngle);
        ctx.translate(-250, -250);
        drawWheel(events);
        ctx.restore();

        if (progress < 1) {
            requestAnimationFrame(animate);
        } else {
            spinning = false;
            const topIdx = getSectorAtTop();
            console.log('✅ Сектор наверху (по расчёту):', topIdx, events[topIdx]?.name);
            console.log('✅ Целевой сектор (из ответа):', idx, events[idx]?.name);
            if (topIdx !== idx) {
                console.error('❌ НЕСООТВЕТСТВИЕ!');
            } else {
                console.log('✅ Всё совпадает!');
            }

            resultName.textContent = targetEvent.name;
            resultDesc.textContent = targetEvent.description;
            if (targetEvent.image) {
                resultImage.src = targetEvent.image;
                resultImage.style.display = 'block';
            } else {
                resultImage.style.display = 'none';
            }
            resultOverlay.style.display = 'block';
        }
    }
    requestAnimationFrame(animate);
}

function connectWebSocket() {
    const ws = new WebSocket('ws://localhost:3001/ws');
    ws.onopen = () => console.log('WebSocket connected');
    ws.onmessage = (msg) => {
        const data = JSON.parse(msg.data);
        if (data.action === 'spin') {
            if (data.roulette === currentRoulette) {
                spinToEvent(data.event);
            } else {
                console.log(`Переключаемся на рулетку ${data.roulette} из-за входящего спина`);
                loadRouletteEvents(data.roulette).then(() => {
                    spinToEvent(data.event);
                });
            }
        }
    };
    ws.onclose = () => setTimeout(connectWebSocket, 1000);
}

rouletteSelect.addEventListener('change', (e) => {
    const name = e.target.value;
    loadRouletteEvents(name);
});

loadRoulettes();
connectWebSocket();