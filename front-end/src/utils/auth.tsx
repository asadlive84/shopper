// src/utils/auth.ts
export const logout = () => {
    localStorage.removeItem('token');
    window.location.reload(); // Reload the page to reflect logout in all tabs
  };