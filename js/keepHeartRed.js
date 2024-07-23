document.addEventListener('DOMContentLoaded', (event) => {
  document.querySelectorAll("[data-is-liked='true']").forEach(element => {
      let postId = element.getAttribute('data-post-id');
      document.getElementById('like-heart-' + postId).classList.add('red');
  });
  document.querySelectorAll("[data-is-disliked='true']").forEach(element => {
    let postId = element.getAttribute('data-post-id');
    document.getElementById('dislike-heart-' + postId).classList.add('red');
});
});