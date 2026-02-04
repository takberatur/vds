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

    // Admob
    private var admobRewardedAd: RewardedAd? = null
    private var isAdmobLoading = false

    // Unity Ads
    private var isUnityLoading = false
    private var isUnityLoaded = false

    // Start.io
    private var isStartIoLoading = false
    private var isStartIoLoaded = false

    private var startAppAd: StartAppAd? = null

    init {
        startAppAd = StartAppAd(activity)
    }

    /**
     * Load all reward ads providers
     */
    fun loadAd() {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled in config")
            return
        }

        loadAdmobReward()
        loadUnityReward()
        loadStartIoReward()
    }

    /**
     * Show reward ad with auto rotation
     */
    fun showAd(
        onAdShown: ((AdsProvider) -> Unit)? = null,
        onRewarded: ((AdsProvider, Int) -> Unit)? = null,
        onAdClosed: ((AdsProvider) -> Unit)? = null,
        onAdFailed: ((AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            onAdFailed?.invoke(AdsProvider.ADMOB)
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.getRewardPriority()) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobRewardedAd != null) {
                        showAdmobReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        Log.d(TAG, "No reward ads available")
        onAdFailed?.invoke(AdsProvider.ADMOB)
    }

    // ========== ADMOB REWARD ==========

    private fun loadAdmobReward() {
        if (isAdmobLoading) return

        isAdmobLoading = true

        val adRequest = AdRequest.Builder().build()

        AdsConfig.ADMOB_REWARDED_ID?.let {
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
        onAdShown: ((AdsProvider) -> Unit)?,
        onRewarded: ((AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        admobRewardedAd?.apply {
            fullScreenContentCallback = object : FullScreenContentCallback() {
                override fun onAdShowedFullScreenContent() {
                    Log.d(TAG, "Admob reward shown")
                    onAdShown?.invoke(AdsProvider.ADMOB)
                }

                override fun onAdDismissedFullScreenContent() {
                    Log.d(TAG, "Admob reward dismissed")
                    admobRewardedAd = null
                    loadAdmobReward() // Reload
                    onAdClosed?.invoke(AdsProvider.ADMOB)
                }

                override fun onAdFailedToShowFullScreenContent(adError: AdError) {
                    Log.e(TAG, "Admob reward show failed: ${adError.message}")
                    admobRewardedAd = null
                    loadAdmobReward()

                    // Try next provider
                    tryNextProvider(AdsProvider.ADMOB, onAdShown, onRewarded, onAdClosed, onAdFailed)
                }
            }

            show(activity) { rewardItem ->
                Log.d(TAG, "User earned reward: ${rewardItem.amount} ${rewardItem.type}")
                onRewarded?.invoke(AdsProvider.ADMOB, rewardItem.amount)
            }
        } ?: run {
            tryNextProvider(AdsProvider.ADMOB, onAdShown, onRewarded, onAdClosed, onAdFailed)
        }
    }

    // ========== UNITY REWARD ==========

    private fun loadUnityReward() {
        if (isUnityLoading) return

        isUnityLoading = true

        AdsConfig.UNITY_REWARDED_ID?.let {
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
        onAdShown: ((AdsProvider) -> Unit)?,
        onRewarded: ((AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        AdsConfig.UNITY_REWARDED_ID?.let {
            UnityAds.show(
                activity,
                it,
                object : IUnityAdsShowListener {
                    override fun onUnityAdsShowStart(placementId: String) {
                        Log.d(TAG, "Unity reward started")
                        onAdShown?.invoke(AdsProvider.UNITY)
                    }


                    override fun onUnityAdsShowComplete(
                        placementId: String,
                        state: UnityAds.UnityAdsShowCompletionState
                    ) {
                        Log.d(TAG, "Unity reward completed: $state")

                        if (state == UnityAds.UnityAdsShowCompletionState.COMPLETED) {
                            // User watched the full video
                            onRewarded?.invoke(AdsProvider.UNITY, 1)
                        }

                        isUnityLoaded = false
                        loadUnityReward() // Reload
                        onAdClosed?.invoke(AdsProvider.UNITY)
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
                            AdsProvider.UNITY,
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

    // ========== START.IO REWARD ==========

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
        onAdShown: ((AdsProvider) -> Unit)?,
        onRewarded: ((AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {


        startAppAd?.setVideoListener(object : VideoListener {
            override fun onVideoCompleted() {
                Log.d(TAG, "Start.io video completed - user rewarded")
                onRewarded?.invoke(AdsProvider.STARTIO, 1)
            }
        })

        startAppAd?.showAd(object : AdDisplayListener {
            override fun adHidden(ad: Ad) {
                Log.d(TAG, "Start.io reward hidden")
                isStartIoLoaded = false
                loadStartIoReward() // Reload
                onAdClosed?.invoke(AdsProvider.STARTIO)
            }

            override fun adDisplayed(ad: Ad) {
                Log.d(TAG, "Start.io reward displayed")
                onAdShown?.invoke(AdsProvider.STARTIO)
            }

            override fun adClicked(ad: Ad) {
                Log.d(TAG, "Start.io reward clicked")
            }

            override fun adNotDisplayed(ad: Ad) {
                Log.e(TAG, "Start.io reward not displayed")
                isStartIoLoaded = false
                loadStartIoReward()

                // Try next provider
                tryNextProvider(AdsProvider.STARTIO, onAdShown, onRewarded, onAdClosed, onAdFailed)
            }
        })
    }

    // ========== HELPER METHODS ==========

    private fun tryNextProvider(
        failedProvider: AdsProvider,
        onAdShown: ((AdsProvider) -> Unit)?,
        onRewarded: ((AdsProvider, Int) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.getRewardPriority().indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.getRewardPriority().drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobRewardedAd != null) {
                        showAdmobReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityReward(onAdShown, onRewarded, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
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

    /**
     * Check if any reward ad is ready
     */
    fun isAdReady(): Boolean {
        return admobRewardedAd != null || isUnityLoaded || isStartIoLoaded
    }

    /**
     * Destroy ads
     */
    fun destroy() {
        admobRewardedAd = null
    }
}