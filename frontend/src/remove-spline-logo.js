function removeSplineLogo() {
  const viewers = document.querySelectorAll('spline-viewer');
  viewers.forEach(viewer => {
    const tryRemove = () => {
      let root = viewer.shadowRoot;
      if (!root) return;
      const logo = root.getElementById('logo');
      if (logo) logo.remove();
      root.querySelectorAll('[class*="logo"], [href*="spline.design"], a[target="_blank"], img[alt*="Spline"]').forEach(el => el.remove());
      root.querySelectorAll('*').forEach(el => {
        if (el.textContent && el.textContent.match(/Built with Spline/i)) {
          el.remove();
        }
      });
    };
    tryRemove();
    setTimeout(tryRemove, 500);
    setTimeout(tryRemove, 1500);
    setTimeout(tryRemove, 3000);
  });
}

window.addEventListener('DOMContentLoaded', removeSplineLogo);
window.addEventListener('load', removeSplineLogo);
setTimeout(removeSplineLogo, 4000); // Fallback for late rendering
