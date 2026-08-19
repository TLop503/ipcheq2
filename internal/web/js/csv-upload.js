document.getElementById('ip-file').addEventListener('change', function () {
    const display = document.getElementById('file-name-display');
    if (this.files.length > 0) {
        display.textContent = this.files[0].name;
    } else {
        display.textContent = 'Choose CSV file…';
    }
});