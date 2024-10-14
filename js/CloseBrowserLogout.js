let lastActivityTime = Date.now();
let logoutTimeout = 300000; // 5 minutes of inactivity before logout
let checkInterval = 60000; // Check every minute
let alertShown = false;

// Update last activity time on any user interaction
document.addEventListener('mousemove', updateLastActivity);
document.addEventListener('keypress', updateLastActivity);
document.addEventListener('click', updateLastActivity);
document.addEventListener('scroll', updateLastActivity);

function updateLastActivity() {
    lastActivityTime = Date.now();
    alertShown = false;
}

// Check for inactivity periodically
setInterval(checkActivity, checkInterval);

function checkActivity() {
    let inactiveTime = Date.now() - lastActivityTime;
    if (inactiveTime > logoutTimeout - 60000 && !alertShown) { // Show alert 1 minute before logout
        alertShown = true;
        if (confirm("You've been inactive for a while. Click OK to stay logged in, or Cancel to log out.")) {
            updateLastActivity();
        } else {
            logout();
        }
    } else if (inactiveTime > logoutTimeout) {
        logout();
    }
}

// Listen for visibility changes
document.addEventListener('visibilitychange', handleVisibilityChange);

function handleVisibilityChange() {
    if (!document.hidden) {
        // Page is visible again, update last activity
        updateLastActivity();
    }
}

function logout() {
    alert("You're being logged out due to inactivity.");
    fetch('/logout', {
        method: 'POST',
        credentials: 'same-origin'
    }).then(() => {
        window.location.href = '/login';
    }).catch(error => {
        console.error('Logout failed:', error);
    });
}

// Manual logout function
function manualLogout() {
    if (confirm("Are you sure you want to log out?")) {
        logout();
    }
}

// update last_active

// Function to update last activity
function updateLastActivity() {
    fetch('/update-activity', {
        method: 'POST',
        credentials: 'same-origin'
    }).catch(error => console.error('Error updating activity:', error));
}

// Set up event listeners for user activity
['mousemove', 'keypress', 'click', 'scroll'].forEach(eventType => {
    document.addEventListener(eventType, updateLastActivity, { passive: true });
});

// Also update activity periodically (every 5 minutes)
setInterval(updateLastActivity, 5 * 60 * 1000);

// Update activity when the page becomes visible again
document.addEventListener('visibilitychange', function() {
    if (!document.hidden) {
        updateLastActivity();
    }
});