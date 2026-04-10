// Apply system color theme
function applyDefault(){
    const darkThemeMq = window.matchMedia("(prefers-color-scheme: dark)");

    // Check for local storage first
    let chosen_color_scheme = localStorage.getItem('color-mode')
    if (chosen_color_scheme){
        // Skip system check if storage contains the color scheme
        return
    }

    if (darkThemeMq.matches){
        localStorage.setItem('color-mode', 'dark')
    } else{
        localStorage.setItem('color-mode', 'light')
    }
}

// Apply color theme
function applyMode(){
    const savedMode = localStorage.getItem('color-mode')
    if (savedMode === 'dark'){
        document.body.classList.add('dark-mode')
        document.getElementById('modeToggle').checked = true
    } else{
        document.body.classList.remove('dark-mode')
        document.getElementById('modeToggle').checked = false
    }
}

// Update mode when the user toggles checkbox
function changeMode(t){
    if (t.checked){
        document.body.classList.add('dark-mode')
        localStorage.setItem('color-mode', 'dark')
    } else{
        document.body.classList.remove('dark-mode')
        localStorage.setItem('color-mode', 'light')
    }
}

document.addEventListener('DOMContentLoaded', applyDefault());
document.getElementById('modeToggle').addEventListener('change', function(){
    changeMode(this)
})
applyMode()
