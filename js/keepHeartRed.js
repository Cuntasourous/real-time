document.addEventListener('DOMContentLoaded', (event) => {
  
  document.querySelectorAll("[data-is-liked='true']").forEach(element => {
      let postId = element.getAttribute('data-post-id');
      let commentId = element.getAttribute('data-comment-id');

      let likeHeart = document.getElementById('like-heart-' + postId)
      let commentHeart = document.getElementById('clike-heart-' + commentId)
      if (likeHeart){
        likeHeart.classList.add('red');
      } else if (commentHeart){
        commentHeart.classList.add('red');
      }
  });

  document.querySelectorAll("[data-is-disliked='true']").forEach(element => {
    let postId = element.getAttribute('data-post-id');
    let commentId = element.getAttribute('data-comment-id')
    ;
    let dislikeHeart = document.getElementById('dislike-heart-' + postId)
    let commentHeart = document.getElementById('cdislike-heart-' + commentId)


    if (dislikeHeart){
      dislikeHeart.classList.add('red');
    } else if (commentHeart){
      commentHeart.classList.add('red');
    }
  });
});