function getUsers() {
    fetch("http://localhost:8080/api")
    .then((response) => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then((data) => {
        const resultDiv = document.getElementById('result');
        resultDiv.innerHTML = ''; // Clear previous content
        
        // Create a container for all images
        const container = document.createElement('div');
        container.style.display = 'flex';
        container.style.flexWrap = 'wrap';
        container.style.gap = '10px';
        
        data.forEach((element) => {
            const img = document.createElement('img');
            img.src = `http://localhost:8080/images?id=${element.id}`;
            img.alt = `Image ${element.id}`;
            img.style.maxWidth = '200px';
            img.style.height = 'auto';
            
            container.appendChild(img);
        });
        
        resultDiv.appendChild(container);
    })
    .catch((error) => {
        console.error('Error:', error);
        document.getElementById('result').innerHTML = 
            `<p style="color: red;">Error loading images: ${error.message}</p>`;
    });
}