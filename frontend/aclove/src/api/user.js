import { useQuery, useQueryClient } from "@tanstack/vue-query";
import { baseFetch } from "./client.js";
import { queryKeys } from "./queryKeys.js";

const USER_STATUS = {
  ACTIVE: "active",
  BANNED: "banned",
  COOLDOWN: "cooldown",
};

async function fetchCurrentUser() {
  return baseFetch("/users/me");
}

function useCurrentUserQuery(options = {}) {
  return useQuery({
    queryKey: queryKeys.users.me(),
    queryFn: fetchCurrentUser,
    staleTime: 1000 * 60 * 5,
    ...options,
  });
}

function usePrefetchCurrentUser() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.users.me(),
      queryFn: fetchCurrentUser,
      staleTime: 1000 * 60 * 5,
    });
  };
}

function useInvalidateCurrentUser() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.invalidateQueries({
      queryKey: queryKeys.users.me(),
    });
  };
}

export {
  USER_STATUS,
  fetchCurrentUser,
  useCurrentUserQuery,
  usePrefetchCurrentUser,
  useInvalidateCurrentUser,
};
