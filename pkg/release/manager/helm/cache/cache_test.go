package cache

import (
	"testing"
	"k8s.io/helm/pkg/helm"
	"walm/pkg/redis"
	"fmt"
	"time"
	"walm/pkg/k8s/client"
	"sync"
)

func TestHelmCache_Resync(t *testing.T) {
	helmClient := helm.NewClient(helm.Host("172.26.0.5:31225"))
	redisClient := redis.CreateFakeRedisClient()
	kubeClient := client.CreateFakeKubeClient("", "C:/kubernetes/0.5/kubeconfig")

	helmCache := &HelmCache{
		redisClient: redisClient,
		helmClient: helmClient,
		kubeClient: kubeClient,
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := helmCache.Resync()
		if err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	_, err := redisClient.GetClient().HSet(redis.WalmReleasesKey, "test1/test1", "test").Result()
	if err != nil {
		fmt.Println(err)
	}

	wg.Wait()
	test, err := redisClient.GetClient().HGet(redis.WalmReleasesKey, "test1/test1").Result()
	if err != nil {
		fmt.Println(err)
	}

	if test != "" {
		t.Fail()
	}

}

func TestHelmCache_GetCache(t *testing.T) {
	helmClient := helm.NewClient(helm.Host("172.26.0.5:31225"))
	redisClient := redis.CreateFakeRedisClient()
	kubeClient := client.CreateFakeKubeClient("", "C:/kubernetes/0.5/kubeconfig")

	helmCache := &HelmCache{
		redisClient: redisClient,
		helmClient: helmClient,
		kubeClient: kubeClient,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := helmCache.Resync()
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg.Wait()

	releaseCache, err := helmCache.GetReleaseCache("kube-system", "app-manager")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("%v\n", *releaseCache)

	releaseCaches, err := helmCache.GetReleaseCaches("", "", 0)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(len(releaseCaches))

	releaseCaches, err = helmCache.GetReleaseCaches("kube-system", "", 0)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(len(releaseCaches))

	releaseCaches, err = helmCache.GetReleaseCaches("", "*security", 0)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(len(releaseCaches))

	releaseCaches, err = helmCache.GetReleaseCaches("", "", 5)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(len(releaseCaches))

	//TODO only return 1 but has 5
	releaseCaches, err = helmCache.GetReleaseCaches("security", "", 5)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(len(releaseCaches))
}

func TestHelmCache_Tmp(t *testing.T) {
	redisClient := redis.CreateFakeRedisClient()
	res, err := redisClient.GetClient().HGet(redis.WalmReleasesKey, "notexists").Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(res)
	}


	num, err := redisClient.GetClient().HDel(redis.WalmReleasesKey, "notexists").Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(num)
	}


	ok, err := redisClient.GetClient().HSet(redis.WalmReleasesKey, "existtest", "test").Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(ok)
	}


	ok, err = redisClient.GetClient().HSetNX(redis.WalmReleasesKey, "existtest", "test").Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(ok)
	}

	num, err = redisClient.GetClient().HDel(redis.WalmReleasesKey, "existtest").Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(num)
	}

}