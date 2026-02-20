/**
 * HR Dashboard - Theme Switcher Logic
 * Simple light/dark mode toggle using localStorage
 */

const themeToggle = document.getElementById('theme-toggle');
const htmlElement = document.documentElement;

// Initialize theme from localStorage or system preferences
const savedTheme = localStorage.getItem('theme') ||
    (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');

htmlElement.setAttribute('data-theme', savedTheme);
updateToggleIcon(savedTheme);

themeToggle.addEventListener('click', () => {
    const currentTheme = htmlElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';

    htmlElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateToggleIcon(newTheme);
});

function updateToggleIcon(theme) {
    // We assume the button has an inner element for the icon or we change text
    const icon = theme === 'light' ? '🌙' : '☀️';
    themeToggle.innerHTML = icon;
}
