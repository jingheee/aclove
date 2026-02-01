export const queryKeys = {
  users: {
    all: ["users"],
    me: () => [...queryKeys.users.all, "me"],
  },
  posts: {
    all: ["posts"],
    lists: () => [...queryKeys.posts.all, "list"],
    list: (filters) => [...queryKeys.posts.lists(), filters],
    details: () => [...queryKeys.posts.all, "detail"],
    detail: (id) => [...queryKeys.posts.details(), id],
  },
};
