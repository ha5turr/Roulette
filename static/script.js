const canvas = document.getElementById('wheel');
const ctx = canvas.getContext('2d');
const resultOverlay = document.getElementById('result-overlay');
const resultName = document.getElementById('result-name');
const resultDesc = document.getElementById('result-desc');
const resultImage = document.getElementById('result-image');
const rouletteSelect = document.getElementById('roulette-select');

let currentRoulette = null; // имя текущей рулетки
let events = [];
let currentAngle = 0;
let spinning = false;

// Загружаем список рулеток
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

// Загружаем события для выбранной рулетки
async function loadRouletteEvents(name) {
    const response = await fetch(`/api/roulette/${name}`);
    const data = await response.json();
    events = data; // массив событий
    currentRoulette = name;
    drawWheel(events);
    // Скрываем результат при смене рулетки
    resultOverlay.style.display = 'none';
}

// Рисуем колесо
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
        const textX = 250 + 150 * Math.cos(midAngle);
        const textY = 250 + 150 * Math.sin(midAngle);
        ctx.save();
        ctx.translate(textX, textY);
        ctx.rotate(midAngle + Math.PI / 2);
        ctx.fillStyle = '#fff';
        ctx.font = 'bold 14px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(events[i].name, 0, 0);
        ctx.restore();
    }
}

// Анимация вращения
function spinToEvent(targetEvent) {
    if (spinning) return;
    spinning = true;
    resultOverlay.style.display = 'none';

    const idx = events.findIndex(e => e.id === targetEvent.id);
    if (idx === -1) {
        spinning = false;
        return;
    }

    const n = events.length;
    const angleStep = (2 * Math.PI) / n;
    const targetAngle = idx * angleStep + angleStep / 2;
    const extraSpins = 5 + Math.random() * 3;
    const finalAngle = (2 * Math.PI) - targetAngle + (2 * Math.PI) * extraSpins;
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

// WebSocket
function connectWebSocket() {
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onopen = () => console.log('WebSocket connected');
    ws.onmessage = (msg) => {
        const data = JSON.parse(msg.data);
        if (data.action === 'spin') {
            // Проверяем, совпадает ли рулетка с текущей
            if (data.roulette === currentRoulette) {
                spinToEvent(data.event);
            } else {
                // Можно автоматически переключиться на эту рулетку
                // либо проигнорировать. Лучше переключиться, чтобы зритель видел нужную рулетку.
                console.log(`Switching to roulette ${data.roulette} due to incoming spin`);
                // Загружаем события этой рулетки и крутим
                loadRouletteEvents(data.roulette).then(() => {
                    // после загрузки событий запускаем вращение
                    spinToEvent(data.event);
                });
            }
        }
    };
    ws.onclose = () => setTimeout(connectWebSocket, 1000);
}

// Обработчик выбора рулетки
rouletteSelect.addEventListener('change', (e) => {
    const name = e.target.value;
    loadRouletteEvents(name);
});

// Инициализация
loadRoulettes();
connectWebSocket();