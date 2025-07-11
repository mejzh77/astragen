// Функции преобразования данных (теперь чисто JavaScript)
function mapNodesToTree(nodes) {
    return nodes.map(n => ({
        id: n.id,
        name: n.name,
        type: 'node',
        // ... другие поля
    }));
}

function mapProductsToTree(products) {
    return products.map(p => ({
        id: p.id,
        name: p.name, 
        type: 'product',
        // ... другие поля
    }));
}

// Инициализация дерева
document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/tree-data')
        .then(response => response.json())
        .then(data => {
            renderTree(data);
            initTreeBehavior();
        });
});

function renderTree(data) {
    const container = document.getElementById('tree');
    container.innerHTML = buildTreeHTML(data);
}

function buildTreeHTML(items) {
    return `<ul>${
        items.map(item => `
        <li data-id="${item.id}" data-type="${item.type}">
            <span class="toggle">${
                hasChildren(item) ? '+' : ''
            }</span>
            <span class="item-name">${item.name}</span>
            ${
                item.systems ? buildTreeHTML(item.systems) :
                item.nodes ? buildTreeHTML(item.nodes) :  // Убрал mapNodesToTree
                item.products ? buildTreeHTML(item.products) :  // Убрал mapProductsToTree
                item.functionBlocks ? buildTreeHTML(item.functionBlocks) : ''
            }
        </li>`).join('')
    }</ul>`;
}

function hasChildren(item) {
    return (item.systems && item.systems.length > 0) || 
           (item.nodes && item.nodes.length > 0) || 
           (item.products && item.products.length > 0) || 
           (item.functionBlocks && item.functionBlocks.length > 0);
}

function initTreeBehavior() {
    // Обработчик для всего дерева
    document.getElementById('tree').addEventListener('click', function(e) {
        const toggle = e.target.closest('.toggle');
        if (toggle) {
            const li = toggle.closest('li');
            if (!li) return;
            
            const ul = li.querySelector('ul');
            if (ul) {
                ul.style.display = ul.style.display === 'none' ? 'block' : 'none';
                toggle.textContent = ul.style.display === 'none' ? '+' : '-';
            }
            return;
        }
        
        // Обработка клика по элементу
        const itemName = e.target.closest('.item-name');
        if (itemName) {
            const li = itemName.closest('li');
            if (!li) return;
            
            const itemId = li.dataset.id;
            const itemType = li.dataset.type;
            
            if (!itemId || !itemType) {
                console.error('Missing data attributes on:', li);
                return;
            }

            console.log(`Loading details for ${itemType} ${itemId}`);
            loadDetails(itemType, itemId);
        }
    });
}

function loadDetails(type, id) {
    console.log(`Fetching details for ${type}/${id}`);
    fetch(`/api/details?type=${type}&id=${id}`)
        .then(response => {
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            return response.json();
        })
        .then(data => {
            document.getElementById('details-content').innerHTML = 
                renderDetails(data, type);
        })
        .catch(error => {
            console.error('Error:', error);
            document.getElementById('details-content').innerHTML = `
                <div class="alert alert-danger">
                    Ошибка загрузки: ${error.message}
                </div>`;
        });
}

function renderDetails(data, type) {
    let html = `<h3>${data.name || data.tag || 'Элемент'}</h3>`;
    html += `<p><strong>Тип:</strong> ${type}</p>`;
    
    // Добавляем отладочную информацию
    html += `<div class="debug-info">
        <p><strong>Полученные данные:</strong></p>
        <pre>${JSON.stringify(data, null, 2)}</pre>
    </div>`;
    
    // Основные поля
    for (const [key, value] of Object.entries(data)) {
        if (value && typeof value === 'object') continue;
        html += `<p><strong>${key}:</strong> ${value}</p>`;
    }
    
    // Специальные поля для разных типов
    switch(type) {
        case 'system':
            if (data.project) {
                html += `<h4>Проект</h4>
                <p><strong>Название:</strong> ${data.project.name}</p>`;
            }
            break;
            
        case 'functionblock':
            if (data.variables && data.variables.length > 0) {
                html += `<h4>Переменные</h4><ul>`;
                data.variables.forEach(v => {
                    html += `<li>${v.direction}: ${v.signalTag}</li>`;
                });
                html += `</ul>`;
            }
            break;
    }
    
    return html;
}
