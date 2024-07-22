document.addEventListener('DOMContentLoaded', (event) => {
    const likeButtons = document.querySelectorAll('.like-button');

    likeButtons.forEach(button => {
        button.addEventListener('click', (e) => {
            console.log("like button clicked");
        });
    });
});

// function toggleHeartColor(postId, action) {
//     const likeHeart = document.getElementById(`like-heart-${postId}`);
//     const dislikeHeart = document.getElementById(`dislike-heart-${postId}`);

//     if (action === 'like') {
//         likeHeart.classList.add('red');
//         dislikeHeart.classList.remove('red');
//     } else if (action === 'dislike') {
//         dislikeHeart.classList.add('red');
//         likeHeart.classList.remove('red');
//     }
// }
