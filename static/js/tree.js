// Функции преобразования данных
function mapItemsToTree(items, type) {
    return items.map(item => ({
        id: item.id,
        name: item.name,
        type: type,
        // ... другие поля
    }));
}

// Инициализация дерева
document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/tree-data')
        .then(response => response.json())
        .then(data => {
            renderTree(data);
            setTimeout(initTreeBehavior, 0); 
        });
});

// Добавьте проверку данных при рендеринге
function renderTree(data) {
    console.log("Tree data:", JSON.stringify(data, null, 2)); // Проверьте структуру
    const container = document.getElementById('tree');
    container.innerHTML = buildTreeHTML(data);
}
function buildTreeHTML(items) {
    if (!items || !items.length) return '';

    return `
    <ul>
        ${items.map(item => {
            const hasChildren = checkHasChildren(item);
            return `
            <li data-id="${item.id}" data-type="${item.type}">
                <span class="toggle">${hasChildren ? '+' : ''}</span>
                <span class="item-name">${item.name}</span>
                ${hasChildren ? buildChildrenHTML(item) : ''}
            </li>`;
        }).join('')}
    </ul>`;
}

function checkHasChildren(item) {
    return (item.systems && item.systems.length > 0) ||
           (item.nodes && item.nodes.length > 0) ||
           (item.products && item.products.length > 0) ||
           (item.functionBlocks && item.functionBlocks.length > 0);
}

function buildChildrenHTML(item) {
    if (item.systems) {
        return buildTreeHTML(item.systems);
    }
    
    // Для систем сначала показываем категории "Узлы" и "Изделия"
    if (item.type === 'system') {
        return `
        <ul>
            ${item.nodes ? `
            <li data-type="category" data-category="nodes">
                <span class="toggle">+</span>
                <span class="item-name">Узлы</span>
                ${buildTreeHTML(item.nodes)}
            </li>` : ''}
            
            ${item.products ? `
            <li data-type="category" data-category="products">
                <span class="toggle">+</span>
                <span class="item-name">Изделия</span>
                ${buildTreeHTML(item.products)}
            </li>` : ''}
        </ul>`;
    }
    
    // Для остальных случаев
    if (item.nodes) return buildTreeHTML(item.nodes);
    if (item.products) return buildTreeHTML(item.products);
    if (item.functionBlocks) return buildTreeHTML(item.functionBlocks);
    
    return '';
}
function buildSystemHTML(system) {
    let html = '';
    if (system.nodes || system.products) {
        html += '<ul>';
        if (system.nodes) {
            html += `
            <li data-type="nodes-category">
                <span class="toggle">+</span>
                <span class="item-name">Узлы</span>
                ${buildTreeHTML(mapItemsToTree(system.nodes, 'node'))}
            </li>`;
        }
        if (system.products) {
            html += `
            <li data-type="products-category">
                <span class="toggle">+</span>
                <span class="item-name">Изделия</span>
                ${buildTreeHTML(mapItemsToTree(system.products, 'product'))}
            </li>`;
        }
        html += '</ul>';
    }
    return html;
}

function buildNodesProductsHTML(item) {
    let html = '';
    if (item.nodes) {
        html += buildTreeHTML(mapItemsToTree(item.nodes, 'node'));
    }
    if (item.products) {
        html += buildTreeHTML(mapItemsToTree(item.products, 'product'));
    }
    return html;
}

function hasChildren(item) {
    return (item.systems && item.systems.length > 0) ||
           (item.nodes && item.nodes.length > 0) ||
           (item.products && item.products.length > 0) ||
           (item.functionBlocks && item.functionBlocks.length > 0);
}

function initTreeBehavior() {
    const treeContainer = document.getElementById('tree');
    if (!treeContainer) return;

    // Инициализация - сворачиваем все поддеревья
    treeContainer.querySelectorAll('ul ul').forEach(ul => {
        ul.style.display = 'none';
    });

    // Обработчик кликов
    treeContainer.addEventListener('click', function(e) {
        const toggle = e.target.closest('.toggle');
        if (toggle) {
            const li = toggle.closest('li');
            if (!li) return;
            
            const ul = li.querySelector('ul');
            if (ul) {
                const isHidden = window.getComputedStyle(ul).display === 'none';
                console.log(ul.style.display); 
                // Применяем изменения
                ul.style.display = isHidden ? 'block' : 'none';
                toggle.textContent = isHidden ? '-' : '+';
            }
            e.stopPropagation();
            return;
        }

        // Обработка кликов по элементам (если нужно)
        const itemName = e.target.closest('.item-name');
        if (itemName) {
            const li = itemName.closest('li');
            const itemId = li?.dataset?.id;
            const itemType = li?.dataset?.type;
            
            if (itemId && itemType && itemType !== 'category') {
                loadDetails(itemType, itemId);
            }
        }
    });
}
function loadDetails(type, id) {
    fetch(`/api/details?type=${type}&id=${id}`)
        .then(response => {
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            return response.json();
        })
        .then(data => {
            document.getElementById('details-content').innerHTML = 
                renderDetailsTable(data, type);
        })
        .catch(error => {
            console.error('Error:', error);
            document.getElementById('details-content').innerHTML = `
                <div class="alert alert-danger">
                    Ошибка загрузки: ${error.message}
                </div>`;
        });
}

function renderDetailsTable(data, type) {
    let html = `<h3>${data.name || data.tag || 'Элемент'}</h3>`;
    html += `<p><strong>Тип:</strong> ${type}</p>`;
    
    // Таблица с основными свойствами
    html += `<table class="details-table">
        <thead>
            <tr>
                <th>Свойство</th>
                <th>Значение</th>
            </tr>
        </thead>
        <tbody>`;
    
    // Основные поля
    for (const [key, value] of Object.entries(data)) {
        if (value && typeof value === 'object') continue;
        html += `<tr>
            <td><strong>${key}</strong></td>
            <td>${value}</td>
        </tr>`;
    }
    
    html += `</tbody></table>`;
    
    // Специальные секции для разных типов
    switch(type) {
        case 'system':
            if (data.project) {
                html += `<h4>Проект</h4>
                <table class="details-table">
                    <tr>
                        <td><strong>Название</strong></td>
                        <td>${data.project.name}</td>
                    </tr>
                </table>`;
            }
            break;
            
        case 'functionblock':
            if (data.variables && data.variables.length > 0) {
                html += `<h4>Переменные</h4>
                <table class="details-table">
                    <thead>
                        <tr>
                            <th>Направление</th>
                            <th>Сигнал</th>
                        </tr>
                    </thead>
                    <tbody>`;
                data.variables.forEach(v => {
                    html += `<tr>
                        <td>${v.direction}</td>
                        <td>${v.signalTag}</td>
                    </tr>`;
                });
                html += `</tbody></table>`;
            }
            break;
    }
    
    return html;
}
