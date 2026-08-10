import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/daily',
    name: 'daily',
    component: () => import('../views/Daily.vue')
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('../views/Search.vue')
  },
  {
    path: '/playlist/:id',
    name: 'playlist',
    component: () => import('../views/Playlist.vue')
  },
  {
    path: '/album/:id',
    name: 'album',
    component: () => import('../views/Album.vue')
  },
  {
    path: '/artist/:id',
    name: 'artist',
    component: () => import('../views/Artist.vue')
  },
  {
    path: '/ranking',
    name: 'ranking',
    component: () => import('../views/Ranking.vue')
  },
  {
    path: '/me',
    name: 'me',
    component: () => import('../views/Me.vue')
  },
  {
    path: '/liked',
    name: 'liked',
    component: () => import('../views/Liked.vue')
  },
  {
    path: '/recent',
    name: 'recent',
    component: () => import('../views/Recent.vue')
  },
  {
    path: '/local',
    name: 'local',
    component: () => import('../views/Local.vue')
  },
  {
    path: '/fm',
    name: 'fm',
    component: () => import('../views/FM.vue')
  },
  {
    path: '/cloud',
    name: 'cloud',
    component: () => import('../views/Cloud.vue')
  },
  {
    path: '/lyric',
    name: 'lyric',
    component: () => import('../views/LyricView.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router