package com.agcforge.videodownloader.helper

import android.app.Activity
import android.util.Log
import com.google.android.gms.ads.AdError
import com.google.android.gms.ads.AdRequest
import com.google.android.gms.ads.FullScreenContentCallback
import com.google.android.gms.ads.LoadAdError
import com.google.android.gms.ads.rewarded.RewardedAd
import com.google.android.gms.ads.rewarded.RewardedAdLoadCallback
import com.startapp.sdk.adsbase.Ad
import com.startapp.sdk.adsbase.StartAppAd
import com.startapp.sdk.adsbase.adlisteners.AdDisplayListener
import com.startapp.sdk.adsbase.adlisteners.AdEventListener
import com.startapp.sdk.adsbase.adlisteners.VideoListener
import com.startapp.sdk.adsbase.model.AdPreferences
import com.unity3d.ads.IUnityAdsLoadListener
import com.unity3d.ads.IUnityAdsShowListener
import com.unity3d.ads.UnityAds

class RewardAdsHelper(private val activity: Activity) {

    private val TAG = "RewardAdsHelper"

    private var admobRewardedAd: RewardedAd? = null
    private var isAdmobLoading = false

    private var isUnityLoading = false
    private var isUnityLoaded = false

    private var isStartIoLoading = false
    private var isStartIoLoaded = false

    private var startAppAd: StartAppAd? = null

    init {
        startAppAd = StartAppAd(activity)
    }

    fun loadAd() {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled in config")
            return
        }

        loadAdmobReward()
        loadUnityReward()
        loadStartIoReward()
    }

    fun showAd(
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onRewarded: ((AdsConfig.AdsProvider, Int) -> Unit)? = null,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            onAdFailed?.invoke(AdsConfig.AdsProvider.ADMOB)
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.REWARD_PRIORITY) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobRewardedAd != null) {
                        showAdmobReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        Log.d(TAG, "No reward ads available")
        onAdFailed?.invoke(AdsConfig.AdsProvider.ADMOB)
    }

    private fun loadAdmobReward() {
        if (isAdmobLoading) return

        isAdmobLoading = true

        val adRequest = AdRequest.Builder().build()

        AdsConfig.admobConfig.rewardedId?.let {
            RewardedAd.load(
                activity,
                it,
                adRequest,
                object : RewardedAdLoadCallback() {
                    override fun onAdLoaded(rewardedAd: RewardedAd) {
                        Log.d(TAG, "Admob reward loaded")
                        admobRewardedAd = rewardedAd
                        isAdmobLoading = false
                    }

                    override fun onAdFailedToLoad(loadAdError: LoadAdError) {
                        Log.e(TAG, "Admob reward failed: ${loadAdError.message}")
                        admobRewardedAd = null
                        isAdmobLoading = false
                    }
                }
            )
        }
    }

    private fun showAdmobReward(
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onRewarded: ((AdsConfig.AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        admobRewardedAd?.apply {
            fullScreenContentCallback = object : FullScreenContentCallback() {
                override fun onAdShowedFullScreenContent() {
                    Log.d(TAG, "Admob reward shown")
                    onAdShown?.invoke(AdsConfig.AdsProvider.ADMOB)
                }

                override fun onAdDismissedFullScreenContent() {
                    Log.d(TAG, "Admob reward dismissed")
                    admobRewardedAd = null
                    loadAdmobReward() // Reload
                    onAdClosed?.invoke(AdsConfig.AdsProvider.ADMOB)
                }

                override fun onAdFailedToShowFullScreenContent(adError: AdError) {
                    Log.e(TAG, "Admob reward show failed: ${adError.message}")
                    admobRewardedAd = null
                    loadAdmobReward()

                    // Try next provider
                    tryNextProvider(AdsConfig.AdsProvider.ADMOB, onAdShown, onRewarded, onAdClosed, onAdFailed)
                }
            }

            show(activity) { rewardItem ->
                Log.d(TAG, "User earned reward: ${rewardItem.amount} ${rewardItem.type}")
                onRewarded?.invoke(AdsConfig.AdsProvider.ADMOB, rewardItem.amount)
            }
        } ?: run {
            tryNextProvider(AdsConfig.AdsProvider.ADMOB, onAdShown, onRewarded, onAdClosed, onAdFailed)
        }
    }

    private fun loadUnityReward() {
        if (isUnityLoading) return

        isUnityLoading = true

        AdsConfig.unityConfig.rewardPlacement?.let {
            UnityAds.load(
                it,
                object : IUnityAdsLoadListener {
                    override fun onUnityAdsAdLoaded(placementId: String) {
                        Log.d(TAG, "Unity reward loaded")
                        isUnityLoaded = true
                        isUnityLoading = false
                    }

                    override fun onUnityAdsFailedToLoad(
                        placementId: String,
                        error: UnityAds.UnityAdsLoadError,
                        message: String
                    ) {
                        Log.e(TAG, "Unity reward failed: $message")
                        isUnityLoaded = false
                        isUnityLoading = false
                    }
                }
            )
        }
    }

    private fun showUnityReward(
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onRewarded: ((AdsConfig.AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        AdsConfig.unityConfig.rewardPlacement?.let { placementId ->
            UnityAds.show(
                activity,
                placementId,
                object : IUnityAdsShowListener {
                    override fun onUnityAdsShowStart(placementId: String) {
                        Log.d(TAG, "Unity reward started")
                        onAdShown?.invoke(AdsConfig.AdsProvider.UNITY)
                    }


                    override fun onUnityAdsShowComplete(
                        placementId: String,
                        state: UnityAds.UnityAdsShowCompletionState
                    ) {
                        Log.d(TAG, "Unity reward completed: $state")

                        if (state == UnityAds.UnityAdsShowCompletionState.COMPLETED) {
                            // User watched the full video
                            onRewarded?.invoke(AdsConfig.AdsProvider.UNITY, 1)
                        }

                        isUnityLoaded = false
                        loadUnityReward() // Reload
                        onAdClosed?.invoke(AdsConfig.AdsProvider.UNITY)
                    }

                    override fun onUnityAdsShowFailure(
                        placementId: String,
                        error: UnityAds.UnityAdsShowError,
                        message: String
                    ) {
                        Log.e(TAG, "Unity reward show failed: $message")
                        isUnityLoaded = false
                        loadUnityReward()

                        // Try next provider
                        tryNextProvider(
                            AdsConfig.AdsProvider.UNITY,
                            onAdShown,
                            onRewarded,
                            onAdClosed,
                            onAdFailed
                        )
                    }

                    override fun onUnityAdsShowClick(placementId: String) {
                        Log.d(TAG, "Unity reward clicked")
                    }
                }
            )
        }
    }

    private fun loadStartIoReward() {
        if (isStartIoLoading) return

        isStartIoLoading = true

        // Start.io uses video ads for reward
        val adPreferences = AdPreferences()

        startAppAd?.loadAd(
            StartAppAd.AdMode.REWARDED_VIDEO,
            object : AdEventListener {
                override fun onReceiveAd(ad: Ad) {
                    Log.d(TAG, "Start.io reward loaded")
                    isStartIoLoaded = true
                    isStartIoLoading = false
                }

                override fun onFailedToReceiveAd(ad: Ad?) {
                    Log.e(TAG, "Start.io reward failed")
                    isStartIoLoaded = false
                    isStartIoLoading = false
                }
            }
        )
    }

    private fun showStartIoReward(
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onRewarded: ((AdsConfig.AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {


        startAppAd?.setVideoListener(object : VideoListener {
            override fun onVideoCompleted() {
                Log.d(TAG, "Start.io video completed - user rewarded")
                onRewarded?.invoke(AdsConfig.AdsProvider.STARTIO, 1)
            }
        })

        startAppAd?.showAd(object : AdDisplayListener {
            override fun adHidden(ad: Ad) {
                Log.d(TAG, "Start.io reward hidden")
                isStartIoLoaded = false
                loadStartIoReward() // Reload
                onAdClosed?.invoke(AdsConfig.AdsProvider.STARTIO)
            }

            override fun adDisplayed(ad: Ad) {
                Log.d(TAG, "Start.io reward displayed")
                onAdShown?.invoke(AdsConfig.AdsProvider.STARTIO)
            }

            override fun adClicked(ad: Ad) {
                Log.d(TAG, "Start.io reward clicked")
            }

            override fun adNotDisplayed(ad: Ad) {
                Log.e(TAG, "Start.io reward not displayed")
                isStartIoLoaded = false
                loadStartIoReward()

                // Try next provider
                tryNextProvider(
                    AdsConfig.AdsProvider.STARTIO,
                    onAdShown,
                    onRewarded,
                    onAdClosed,
                    onAdFailed
                )
            }
        })
    }

    private fun tryNextProvider(
        failedProvider: AdsConfig.AdsProvider,
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onRewarded: ((AdsConfig.AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.REWARD_PRIORITY.indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.REWARD_PRIORITY.drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobRewardedAd != null) {
                        showAdmobReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        // All providers failed
        Log.e(TAG, "All reward providers failed")
        onAdFailed?.invoke(failedProvider)
    }


    fun isAdReady(): Boolean {
        return admobRewardedAd != null || isUnityLoaded || isStartIoLoaded
    }


    fun destroy() {
        admobRewardedAd = null
    }
}